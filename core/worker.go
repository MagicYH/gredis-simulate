package core

import (
	"bufio"
	"bytes"
	"context"
	"gredissimulate/core/processor"
	"gredissimulate/core/proto"
	"gredissimulate/logger"
	"net"
	"reflect"
)

// Worker : worker for client
type Worker struct {
	ctx         context.Context
	conn        net.Conn
	reader      *bufio.Reader
	newProcFunc func() processor.Processor
}

// NewWorker : Create new worker instance
func NewWorker(ctx context.Context, conn net.Conn, function func() processor.Processor) (*Worker, error) {
	worker := &Worker{
		ctx:         ctx,
		conn:        conn,
		reader:      bufio.NewReader(conn),
		newProcFunc: function,
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
		// proc := &processor.SimpleProc{}
		proc := worker.newProcFunc()
		err := worker.ProcessMultiCmd(proc)
		if nil != err {
			break
		}
	}
}

// ProcessMultiCmd : Process multi commands
func (worker *Worker) ProcessMultiCmd(proc processor.Processor) error {
	for {
		parser := proto.NewParser()
		request, err := parser.ParseCmd(worker)

		var responseGroup *proto.ResponseGroup
		if nil != err {
			logger.LogError("Parse cmd fail", err)
			if "proto.NetError" == reflect.TypeOf(err).String() {
				return err
			}
			responseGroup.AppendResponse(proto.NewErrorRes("Parse cmd fail"))
		} else {
			// Use processor
			responseGroup, err = processor.ProcessReq(proc, request)
			if nil != err {
				logger.LogError(err)
			}
		}

		worker.conn.Write(proto.BuildMultiResBinary(responseGroup))

		if !proc.IsMulti() {
			break
		}
	}

	return nil
}

// ReadLine : Readline of stream from socket
func (worker *Worker) ReadLine() (string, error) {
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
