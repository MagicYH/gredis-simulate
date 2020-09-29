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
)

// Worker : worker for client
type Worker struct {
	ctx         context.Context
	conn        net.Conn
	newProcFunc processor.Create
	needAuth    bool
	passwd      string
	scanner     *bufio.Scanner
	readOnly    bool
	slaveModel  bool
}

// WorkerConf : worker config
type WorkerConf struct {
	Passwd      string
	ReadOnly    bool
	NewProcFunc processor.Create
}

// NewWorker : Create new worker instance
func NewWorker(ctx context.Context, conn net.Conn, conf WorkerConf) (*Worker, error) {
	passwd := conf.Passwd
	needAuth := false
	if "" != conf.Passwd {
		needAuth = true
	}
	worker := &Worker{
		ctx:         ctx,
		conn:        conn,
		newProcFunc: conf.NewProcFunc,
		needAuth:    needAuth,
		passwd:      passwd,
		scanner:     bufio.NewScanner(conn),
		readOnly:    conf.ReadOnly,
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
	defer func() {
		logger.LogInfo("Remote client disconnect: ", worker.conn.RemoteAddr())
		worker.conn.Close()
	}()

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
		var response *proto.Response
		if nil != err {
			if "proto.NetError" == reflect.TypeOf(err).String() {
				return err
			}

			response = proto.NewErrorRes("Parse cmd fail")
		} else {
			if worker.NeedAuth() {
				if "AUTH" == request.Cmd {
					response, err = proc.AUTH(request)

					if nil == err {
						worker.needAuth = false
					} else {
						worker.needAuth = true
					}
				} else {
					response = proto.NewErrorRes("NOAUTH Authentication required.")
				}
			} else {
				// Use processor
				response, err = processor.ProcessReq(proc, request)
				if nil != err {
					logger.LogError(err)
				}
			}
		}

		if false == worker.readOnly {
			worker.conn.Write([]byte(proto.BuildResBinary(response)))
		}

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
