package core

import (
	"bufio"
	"context"
	"errors"
	"gredissimulate/core/processor"
	"gredissimulate/core/proto"
	"gredissimulate/logger"
	"net"
	"reflect"
	"strings"
)

// Worker : worker for client
type Worker struct {
	ctx         context.Context
	conn        net.Conn
	newProcFunc processor.Create
	needAuth    bool
	passwd      string
	scanner     *bufio.Scanner
}

// NewWorker : Create new worker instance
func NewWorker(ctx context.Context, conn net.Conn, conf ServerConf, function processor.Create) (*Worker, error) {
	passwd := conf.Passwd
	needAuth := false
	if "" != conf.Passwd {
		needAuth = true
	}
	worker := &Worker{
		ctx:         ctx,
		conn:        conn,
		newProcFunc: function,
		needAuth:    needAuth,
		passwd:      passwd,
		scanner:     bufio.NewScanner(conn),
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
		proc := worker.newProcFunc(worker.passwd)
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

		responseGroup := proto.NewResponseGroup()
		if nil != err {
			logger.LogError("Parse cmd fail", err)
			if "proto.NetError" == reflect.TypeOf(err).String() {
				return err
			}
			responseGroup.AppendResponse(proto.NewErrorRes("Parse cmd fail"))
		} else {
			if worker.NeedAuth() {
				if "Auth" == strings.Title(request.Cmd) {
					response, err := proc.Auth(request)
					responseGroup.AppendResponse(response)

					if nil == err {
						worker.needAuth = false
					} else {
						worker.needAuth = true
					}
				} else {
					responseGroup = &proto.ResponseGroup{}
					responseGroup.AppendResponse(proto.NewErrorRes("NOAUTH Authentication required."))
				}
			} else {
				// Use processor
				responseGroup, err = processor.ProcessReq(proc, request)
				if nil != err {
					logger.LogError(err)
				}
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
func (worker *Worker) ReadLine() (content string, err error) {
	if worker.scanner.Scan() {
		content = worker.scanner.Text()
	} else {
		err = worker.scanner.Err()
		if nil == err {
			err = errors.New("EOF")
		}
	}
	return
}

// NeedAuth : is worker need auth
func (worker *Worker) NeedAuth() bool {
	return worker.needAuth && "" != worker.passwd
}
