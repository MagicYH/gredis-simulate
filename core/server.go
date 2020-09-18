package core

import (
	"context"
	"errors"
	"gredissimulate/core/processor"
	"gredissimulate/logger"
	"net"
	"strconv"
)

// ServerConf : Configure of server
type ServerConf struct {
	Port int
}

// Server : server
type Server struct {
	ctx      context.Context
	conf     ServerConf
	listener net.Listener
}

// NewServer : Create new server
func NewServer(conf ServerConf) (*Server, error) {

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
	if nil != err {
		return nil, errors.New("Create server fail: " + err.Error())
	}

	server := &Server{
		conf:     conf,
		listener: listener,
	}
	return server, nil
}

// Start : Start server
func (server *Server) Start(ctx context.Context) error {
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
	worker, err := NewWorker(ctx, conn, processor.NewSimpleProc)
	if nil != err {
		logger.LogError("Create new worker fail: " + err.Error())
		return
	}

	go worker.DoServe()
}
