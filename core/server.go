package core

import (
	"bufio"
	"context"
	"errors"
	"gredissimulate/core/processor"
	"gredissimulate/core/proto"
	"gredissimulate/logger"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MagicYH/rdb"
	"github.com/MagicYH/rdb/nopdecoder"
)

// ServerConf : Configure of server
type ServerConf struct {
	Port    int
	Passwd  string
	SlaveOf string
}

// Server : server
type Server struct {
	ctx         context.Context
	conf        ServerConf
	listener    net.Listener
	newProcFunc processor.Create
	runid       string
	offset      int
}

// NewServer : Create new server
//
// @param conf ServerConf : Server config, etc: Listen port
// @param function func() processor.Processor : Function that create a new processor instance
func NewServer(conf ServerConf, function processor.Create) (*Server, error) {

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
	if nil != err {
		return nil, errors.New("Create server fail: " + err.Error())
	}

	server := &Server{
		conf:        conf,
		listener:    listener,
		newProcFunc: function,
	}
	return server, nil
}

// Start : Start server
func (server *Server) Start(ctx context.Context) error {
	if "" != server.conf.SlaveOf {
		go server.doSync()
	}

	server.ctx = ctx
	for {
		select {
		case <-server.ctx.Done():
			return nil
		default:
			conn, err := server.listener.Accept()
			if nil == err {
				logger.LogInfo("New connection from: ", conn.RemoteAddr())
				server.handle(conn)
			} else {
				logger.LogError("Accept conn fail: " + err.Error())
			}
		}
	}
}

// Close : Close server
func (server *Server) Close() error {
	return server.listener.Close()
}

func (server *Server) handle(conn net.Conn) {
	ctx, _ := context.WithCancel(server.ctx)
	conf := WorkerConf{
		Passwd:      server.conf.Passwd,
		NewProcFunc: server.newProcFunc,
		ReadOnly:    false,
	}
	worker, err := NewWorker(ctx, conn, conf)
	if nil != err {
		logger.LogError("Create new worker fail: " + err.Error())
		return
	}

	go worker.DoServe()
}

func (server *Server) doSync() {
	ctx, _ := context.WithCancel(server.ctx)
	for {
		select {
		case <-ctx.Done():
			break
		default:
			func() {
				conn, err := net.DialTimeout("tcp", server.conf.SlaveOf, 3*time.Second)
				defer conn.Close()
				if nil != err {
					logger.LogError("Fail to sync from slave", server.conf.SlaveOf, ", error:", err)
					os.Exit(255)
				}
				logger.LogInfo("Create new connect to", server.conf.SlaveOf)

				psyncCmd := server.getPsyncCmd()
				logger.LogInfo("Send psync cmd to master", psyncCmd)
				_, err = conn.Write([]byte(psyncCmd))
				if nil != err {
					logger.LogError("Send full sync message fail", err)
					return
				}
				reader := bufio.NewReader(conn)
				isFull, err := server.getSyncBaseInfo(reader)
				if nil != err {
					logger.LogError("Get full sync base info error:", err)
					return
				}

				if isFull {
					err = server.fullSync(conn)
				}
				if nil != err {
					return
				}
				server.continueSync(conn)
			}()
		}
	}

	// conn, err := net.DialTimeout("tcp", server.conf.SlaveOf, 3*time.Second)
	// if nil != err {
	// 	logger.LogError("Fail to sync from slave", server.conf.SlaveOf, ", error:", err)
	// 	os.Exit(255)
	// }

	// err = server.fullSync(conn)
	// if nil != err {
	// 	logger.LogError("Do full sync error", err)
	// }

	// go server.continueSync(conn)
}

func (server *Server) fullSync(conn net.Conn) error {
	logger.LogInfo("Full sync from master")
	reader := bufio.NewReader(conn)
	length, err := getRdbLength(reader)
	if nil != err {
		logger.LogError("Get rdb file length error")
		return err
	}

	rdbPath := "temp.rdb"
	err = dumpRdbToLocal(reader, length, rdbPath)
	if nil != err {
		logger.LogError("Save rdb file error", err)
		return err
	}
	err = server.loadRdbFile(rdbPath)
	if nil != err {
		logger.LogError("Load rdb file error", err)
		return err
	}
	return nil
}

func (server *Server) continueSync(conn net.Conn) {
	logger.LogInfo("Begin continue sync servce")
	ctx, _ := context.WithCancel(server.ctx)
	conf := WorkerConf{
		Passwd:      "",
		NewProcFunc: server.newProcFunc,
		ReadOnly:    true,
	}
	worker, err := NewWorker(ctx, conn, conf)
	if nil != err {
		logger.LogError("Create new worker fail: " + err.Error())
		return
	}
	worker.DoServe()

	server.offset = server.offset + worker.readBytes
}

func (server *Server) getSyncBaseInfo(reader *bufio.Reader) (isFull bool, err error) {
	isFull = true
	var content string
	content, err = reader.ReadString('\n')
	if nil != err {
		return
	}

	content = strings.Trim(content, "\r\n")
	strs := strings.Split(content, " ")

	logger.LogInfo("psync base info:", strings.Trim(content, "\r\n"))
	if "+FULLRESYNC" == strs[0] {
		if len(strs) == 3 {
			server.runid = strs[1]
			server.offset, err = strconv.Atoi(strs[2])
		} else {
			err = errors.New("Get full sync info error, wrong content: " + content)
		}
	} else if "+CONTINUE" == strs[0] {
		isFull = false
	} else {
		err = errors.New("Unexpect sync response")
	}
	return
}

func getRdbLength(reader *bufio.Reader) (length int, err error) {
	content, err := reader.ReadString('\n')
	if nil != err {
		return
	}
	content = strings.Trim(content, "$\r\n")
	length, err = strconv.Atoi(content)
	return
}

func dumpRdbToLocal(reader *bufio.Reader, length int, rdbPath string) error {
	bufferLength := 10240

	var fd *os.File
	var err error
	fd, err = os.OpenFile(rdbPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
	defer fd.Close()
	if nil != err {
		return err
	}
	// read data from socket
	for length > bufferLength {
		buffer, err := reader.Peek(bufferLength)
		if nil != err {
			return err
		}
		reader.Discard(bufferLength)

		err = writeToFile(fd, buffer)
		if nil != err {
			return err
		}

		length = length - bufferLength
	}

	if length <= 0 {
		return nil
	}

	buffer, err := reader.Peek(length)
	reader.Discard(length)
	if nil != err {
		return err
	}
	return writeToFile(fd, buffer)
}

func writeToFile(fd *os.File, buffer []byte) error {
	realLength, err := fd.Write(buffer)
	if nil != err {
		return err
	}
	if len(buffer) != realLength {
		return errors.New("Data write to rdb file length not match")
	}
	return nil
}

type decoder struct {
	nopdecoder.NopDecoder
	proc processor.Processor
}

func (p *decoder) Set(key, value []byte, expiry int64) {
	req := &proto.Request{Cmd: "SET", Params: []string{string(key), string(value)}}
	_, err := processor.ProcessReq(p.proc, req)
	if nil != err {
		logger.LogError("Set data from rdb fail:", err)
	}
}

func (p *decoder) Hset(key, field, value []byte) {
	req := &proto.Request{Cmd: "HSET", Params: []string{string(key), string(field), string(value)}}
	_, err := processor.ProcessReq(p.proc, req)
	if nil != err {
		logger.LogError("Set data from rdb fail:", err)
	}
}

func (server *Server) loadRdbFile(rdbPath string) error {
	f, err := os.Open(rdbPath)
	defer f.Close()
	if nil != err {
		return err
	}
	proc := server.newProcFunc(server.conf.Passwd)
	err = rdb.Decode(f, &decoder{proc: proc})
	return nil
}

func (server *Server) getPsyncCmd() string {
	psyncCmd := "PSYNC "
	if "" == server.runid {
		psyncCmd = psyncCmd + "? -1\n"
	} else {
		psyncCmd = psyncCmd + server.runid + " " + strconv.Itoa(server.offset) + "\n"
	}
	return psyncCmd
}
