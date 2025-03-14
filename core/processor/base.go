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
	EXEC(*proto.Request) (*proto.Response, error)
	IsMulti() bool
	SetMulti(bool)
	GetReqQue() []*proto.Request
	AppendReq(*proto.Request)
}

// ProcessReq : Process request
func ProcessReq(proc Processor, req *proto.Request) (res *proto.Response, err error) {
	cmd := req.Cmd

	v := reflect.ValueOf(proc)
	method := v.MethodByName(cmd)
	if method.IsValid() {
		if !proc.IsMulti() {
			if "EXEC" != cmd {
				result := method.Call([]reflect.Value{reflect.ValueOf(req)})
				if !result[0].IsNil() {
					res = result[0].Interface().(*proto.Response)
				} else {
					res = proto.NewErrorRes("Process `" + cmd + "` error")
				}

				if !result[1].IsNil() {
					err = result[1].Interface().(error)
				}
			} else {
				res, _ = proc.EXEC(req)
			}
		} else {
			if "EXEC" != cmd {
				if "MULTI" != cmd {
					proc.AppendReq(req)
					res = proto.NewResponse(proto.RES_TYPE_STATE)
					res.SetString("QUEUE")
				} else {
					res, _ = proc.MULTI(req)
				}
			} else {
				res, _ = execMulti(proc)
			}
		}
	} else {
		res = proto.NewErrorRes("Unknow command")
	}

	return
}

func execMulti(proc Processor) (res *proto.Response, err error) {
	if proc.IsMulti() {
		proc.SetMulti(false)
		res = proto.NewResponse(proto.RES_TYPE_MULTI)
		for _, request := range proc.GetReqQue() {
			cmd := request.Cmd

			v := reflect.ValueOf(proc)
			method := v.MethodByName(cmd)
			result := method.Call([]reflect.Value{reflect.ValueOf(request)})
			if !result[0].IsNil() {
				res.SetResponse(result[0].Interface().(*proto.Response))
			} else {
				res.SetResponse(proto.NewErrorRes("Process `" + cmd + "` error"))
			}
		}
	} else {
		res = proto.NewErrorRes("EXEC without MULTI")
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

// GetReqQue : Get request que from processor
func (proc *BaseProc) GetReqQue() []*proto.Request {
	return proc.reqQue
}

// SetMulti : Update isMulti flag
func (proc *BaseProc) SetMulti(flag bool) {
	proc.isMulti = flag
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
	res.SetString("OK")
	return
}

// PING : Empty processor ping
func (proc *BaseProc) PING(req *proto.Request) (res *proto.Response, err error) {
	res = proto.NewResponse(proto.RES_TYPE_STATE)
	res.SetString("PONG")
	return
}

// AUTH : Empty processor auth
func (proc *BaseProc) AUTH(req *proto.Request) (res *proto.Response, err error) {
	if "" == proc.passwd {
		res = proto.NewErrorRes("ERR Client sent AUTH, but no password is set")
	} else {
		if proc.passwd == req.Params[0] {
			res = proto.NewResponse(proto.RES_TYPE_STATE)
			res.SetString("OK")
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
		res.SetString("OK")
	} else {
		res = proto.NewErrorRes("MULTI calls can not be nested")
	}
	return
}

// EXEC : Empty processor auth
func (proc *BaseProc) EXEC(req *proto.Request) (res *proto.Response, err error) {
	if proc.isMulti {
		proc.isMulti = false
		res = proto.NewResponse(proto.RES_TYPE_MULTI)
		for _, request := range proc.reqQue {
			cmd := request.Cmd
			v := reflect.ValueOf(proc)
			method := v.MethodByName(cmd)
			result := method.Call([]reflect.Value{reflect.ValueOf(request)})
			if !result[0].IsNil() {
				res.SetResponse(result[0].Interface().(*proto.Response))
			} else {
				res.SetResponse(proto.NewErrorRes("Process `" + cmd + "` error"))
			}
		}
	} else {
		res = proto.NewErrorRes("EXEC without MULTI")
	}
	return
}
