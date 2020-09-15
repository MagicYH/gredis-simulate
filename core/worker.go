package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"gredissimulate/core/processor"
	"gredissimulate/core/proto"
	"gredissimulate/logger"
	"net"
	"reflect"
)

// Worker : worker for client
type Worker struct {
	ctx    context.Context
	conn   net.Conn
	reader *bufio.Reader
}

// NewWorker : Create new worker instance
func NewWorker(ctx context.Context, conn net.Conn) (*Worker, error) {
	worker := &Worker{
		ctx:    ctx,
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
	return worker, nil
}

// NetError : NetError
type NetError struct {
	s string
}

// NewNetError : Create new NetError
func NewNetError(s string) *NetError {
	return &NetError{
		s: s,
	}
}

func (e *NetError) Error() string {
	return e.s
}

// DoServe : Do the worker's work
func (worker *Worker) DoServe() {
	defer worker.conn.Close()

	for {
		request, err := worker.parseCmd()
		if nil != err {
			logger.LogError("Parse cmd fail", err)
			if "NetError" == reflect.TypeOf(err).String() {
				break
			}
		} else {
			proc := &processor.EmptyProc{}
			response, err := processor.ProcessReq(proc, request)
			if nil != err {
				logger.LogError(err)
			}
			worker.conn.Write(proto.BuildResBinary(response))
		}
	}
}

func (worker *Worker) readLine() (string, error) {
	var buf bytes.Buffer
	for {
		content, prefix, err := worker.reader.ReadLine()
		if nil != err {
			return "", NewNetError(err.Error())
		}
		buf.Write(content)
		if !prefix {
			break
		}
	}
	return buf.String(), nil
}

func (worker *Worker) parseCmd() (*proto.Request, error) {
	parser := proto.NewParser()
	for {
		content, err := worker.readLine()
		if nil != err {
			logger.LogError("Worker read line error: ", err)
			return nil, err
		}

		fmt.Println("Get content", content)
		isOk, err := parser.DoParse(content)
		if nil != err {
			logger.LogInfo("Parse cmd error")
			return nil, err
		}
		if isOk {
			break
		}
	}

	return parser.GetRequest(), nil
}

func (worker *Worker) procRequest(request *proto.Request) (*proto.Response, error) {
	response := &proto.Response{}
	return response, nil
}
