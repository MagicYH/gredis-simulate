package processor

import (
	"gredissimulate/core/proto"
)

var set map[string]string
var hash map[string](map[string]string)

// SimpleProc : SimpleProc
type SimpleProc struct {
	BaseProc
}

// NewSimpleProc : Create new simple processor
func NewSimpleProc(passwd string) Processor {
	return &SimpleProc{BaseProc{passwd: passwd}}
}

// GET : Empty processor get
func (proc *SimpleProc) GET(req *proto.Request) (res *proto.Response, err error) {
	if 1 == len(req.Params) {
		k := req.Params[0]
		res = proto.NewResponse(proto.RES_TYPE_BULK)
		if v, ok := set[k]; ok {
			res.SetString(v)
		}
	} else {
		res = proto.NewErrorRes("wrong number of arguments for 'GET' command")
	}
	return
}

// SET : Empty processor set
func (proc *SimpleProc) SET(req *proto.Request) (res *proto.Response, err error) {
	if len(req.Params) == 2 {
		k := req.Params[0]
		v := req.Params[1]
		set[k] = v
		res = proto.NewResponse(proto.RES_TYPE_STATE)
		res.SetString("OK")
	} else {
		res = proto.NewErrorRes("wrong number of arguments for 'SET' command")
	}
	return
}

// HSET : Empty processor hset
func (proc *SimpleProc) HSET(req *proto.Request) (res *proto.Response, err error) {
	k := req.Params[0]
	var data map[string]string
	var ok bool
	if data, ok = hash[k]; !ok {
		data = make(map[string]string)
	}

	updateCount := 0
	for i := 1; i < len(req.Params); i = i + 2 {
		field := req.Params[i]
		value := req.Params[i+1]
		data[field] = value
		updateCount++
	}
	hash[k] = data
	res = proto.NewResponse(proto.RES_TYPE_INT)
	res.SetInt(updateCount)
	return
}

// HGET : Empty processor hget
func (proc *SimpleProc) HGET(req *proto.Request) (res *proto.Response, err error) {
	k := req.Params[0]
	var data map[string]string
	var ok bool
	if data, ok = hash[k]; !ok {
		data = make(map[string]string)
	}

	res = proto.NewResponse(proto.RES_TYPE_STATE)
	var v string
	if v, ok = data[req.Params[1]]; ok {
		res.SetString(v)
	}

	return
}

// HGETALL : Empty processor hgetall
func (proc *SimpleProc) HGETALL(req *proto.Request) (res *proto.Response, err error) {
	k := req.Params[0]
	var data map[string]string
	var ok bool
	if 1 == len(req.Params) {
		res = proto.NewResponse(proto.RES_TYPE_MULTI)
		if data, ok = hash[k]; ok {
			for field, value := range data {
				r1 := proto.NewResponse(proto.RES_TYPE_BULK)
				r1.SetString(field)
				r2 := proto.NewResponse(proto.RES_TYPE_BULK)
				r2.SetString(value)
				res.SetResponse(r1)
				res.SetResponse(r2)
			}
		}
	} else {
		res = proto.NewErrorRes("wrong number of arguments for 'set' command")
	}
	return
}

// SELECT : SELECT command
func (proc *SimpleProc) SELECT(req *proto.Request) (res *proto.Response, err error) {
	res = proto.NewResponse(proto.RES_TYPE_STATE)
	res.SetString("OK")
	return
}

func init() {
	set = make(map[string]string)
	hash = make(map[string](map[string]string))
}
