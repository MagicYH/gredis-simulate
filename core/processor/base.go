package processor

import (
	"errors"
	"gredissimulate/core/proto"
	"reflect"
)

// Create : construct function define
type Create func(string) Processor

// Processor : Processor interface
type Processor interface {
	PING(*proto.Request) (*proto.Response, error)
	AUTH(*proto.Request) (*proto.Response, error)
	MULTI(*proto.Request) (*proto.Response, error)
	EXEC(*proto.Request) ([]*proto.Response, error)
	IsMulti() bool
	AppendReq(*proto.Request)
}

// ProcessReq : Process request
func ProcessReq(proc Processor, req *proto.Request) (group *proto.ResponseGroup, err error) {
	cmd := req.Cmd

	group = proto.NewResponseGroup()
	v := reflect.ValueOf(proc)
	method := v.MethodByName(cmd)
	if method.IsValid() {
		if !proc.IsMulti() {
			if "EXEC" != cmd {
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
				responses, _ := proc.EXEC(req)
				group.AppendResponse(responses[0])
			}
		} else {
			if "EXEC" != cmd {
				if "MULTI" != cmd {
					proc.AppendReq(req)
					oneResponse := proto.NewResponse(proto.RES_TYPE_STATE)
					oneResponse.SetState("QUEUE")
					group.AppendResponse(oneResponse)
				} else {
					response, _ := proc.MULTI(req)
					group.AppendResponse(response)
				}
			} else {
				group.SetType(proto.RESPONSE_GROUP_MULTI)
				responses, _ := proc.EXEC(req)
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

// GET : Empty processor get
func (proc *BaseProc) GET(req *proto.Request) (res *proto.Response, err error) {
	res = proto.NewResponse(proto.RES_TYPE_BULK)
	return
}

// SET : Empty processor set
func (proc *BaseProc) SET(req *proto.Request) (res *proto.Response, err error) {
	res = proto.NewResponse(proto.RES_TYPE_STATE)
	res.SetState("OK")
	return
}

// PING : Empty processor ping
func (proc *BaseProc) PING(req *proto.Request) (res *proto.Response, err error) {
	res = proto.NewResponse(proto.RES_TYPE_STATE)
	res.SetState("PONG")
	return
}

// AUTH : Empty processor auth
func (proc *BaseProc) AUTH(req *proto.Request) (res *proto.Response, err error) {
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

// MULTI : Empty processor multi
func (proc *BaseProc) MULTI(req *proto.Request) (res *proto.Response, err error) {
	if !proc.isMulti {
		proc.isMulti = true
		res = proto.NewResponse(proto.RES_TYPE_STATE)
		res.SetState("OK")
	} else {
		res = proto.NewErrorRes("MULTI calls can not be nested")
	}
	return
}
