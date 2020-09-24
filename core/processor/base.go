package processor

import (
	"errors"
	"gredissimulate/core/proto"
	"reflect"
	"strings"
)

// Create : construct function define
type Create func(string) Processor

// Processor : Processor interface
type Processor interface {
	Ping(*proto.Request) (*proto.Response, error)
	Auth(*proto.Request) (*proto.Response, error)
	Multi(*proto.Request) (*proto.Response, error)
	Exec(*proto.Request) ([]*proto.Response, error)
	IsMulti() bool
	AppendReq(*proto.Request)
}

// ProcessReq : Process request
func ProcessReq(proc Processor, req *proto.Request) (group *proto.ResponseGroup, err error) {
	cmd := strings.Title(req.Cmd)

	group = proto.NewResponseGroup()
	v := reflect.ValueOf(proc)
	method := v.MethodByName(cmd)
	if method.IsValid() {
		if !proc.IsMulti() {
			if "Exec" != cmd {
				result := method.Call([]reflect.Value{reflect.ValueOf(req)})
				if !result[0].IsNil() {
					group.AppendResponse(result[0].Interface().(*proto.Response))
				} else {
					group.AppendResponse(proto.NewErrorRes("Process `" + cmd + "` error"))
				}

				if !result[1].IsNil() {
					err = result[1].Interface().(error)
				}
			} else {
				responses, _ := proc.Exec(req)
				group.AppendResponse(responses[0])
			}
		} else {
			if "Exec" != cmd {
				if "Multi" != cmd {
					proc.AppendReq(req)
					oneResponse := proto.NewResponse(proto.RES_TYPE_STATE)
					oneResponse.SetState("QUEUE")
					group.AppendResponse(oneResponse)
				} else {
					response, _ := proc.Multi(req)
					group.AppendResponse(response)
				}
			} else {
				group.SetType(proto.RESPONSE_GROUP_MULTI)
				responses, _ := proc.Exec(req)
				for _, response := range responses {
					group.AppendResponse(response)
				}
			}
		}
	} else {
		group.AppendResponse(proto.NewErrorRes("Unknow command"))
	}

	return
}

// BaseProc : Do nothing
type BaseProc struct {
	isMulti bool
	passwd  string
	reqQue  []*proto.Request
}

// IsCmdSupport : Whether cmd support by processor
func (proc *BaseProc) IsCmdSupport(cmd string) bool {
	v := reflect.ValueOf(proc)
	method := v.MethodByName(cmd)
	return method.IsValid()
}

// IsMulti : Is processor in multi processing
func (proc *BaseProc) IsMulti() bool {
	return proc.isMulti
}

// AppendReq : Push request to proc queue
func (proc *BaseProc) AppendReq(req *proto.Request) {
	proc.reqQue = append(proc.reqQue, req)
}

// Get : Empty processor get
func (proc *BaseProc) Get(req *proto.Request) (res *proto.Response, err error) {
	res = proto.NewResponse(proto.RES_TYPE_BULK)
	return
}

// Set : Empty processor set
func (proc *BaseProc) Set(req *proto.Request) (res *proto.Response, err error) {
	res = proto.NewResponse(proto.RES_TYPE_STATE)
	res.SetState("OK")
	return
}

// Ping : Empty processor ping
func (proc *BaseProc) Ping(req *proto.Request) (res *proto.Response, err error) {
	res = proto.NewResponse(proto.RES_TYPE_STATE)
	res.SetState("PONG")
	return
}

// Auth : Empty processor auth
func (proc *BaseProc) Auth(req *proto.Request) (res *proto.Response, err error) {
	if "" == proc.passwd {
		res = proto.NewErrorRes("ERR Client sent AUTH, but no password is set")
	} else {
		if proc.passwd == req.Params[0] {
			res = proto.NewResponse(proto.RES_TYPE_STATE)
			res.SetState("OK")
		} else {
			res = proto.NewErrorRes("ERR invalid password")
			err = errors.New("ERR invalid password")
		}
	}
	return
}

// Multi : Empty processor multi
func (proc *BaseProc) Multi(req *proto.Request) (res *proto.Response, err error) {
	if !proc.isMulti {
		proc.isMulti = true
		res = proto.NewResponse(proto.RES_TYPE_STATE)
		res.SetState("OK")
	} else {
		res = proto.NewErrorRes("MULTI calls can not be nested")
	}
	return
}
