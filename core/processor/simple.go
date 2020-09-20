package processor

import (
	"gredissimulate/core/proto"
	"reflect"
	"strconv"
	"strings"
)

var set map[string]string
var hash map[string](map[string]string)

// SimpleProc : SimpleProc
type SimpleProc struct {
	BaseProc
}

// NewSimpleProc : Create new simple processor
func NewSimpleProc() Processor {
	return &SimpleProc{}
}

// Get : Empty processor get
func (proc *SimpleProc) Get(req *proto.Request) (res *proto.Response, err error) {
	if 1 == len(req.Params) {
		k := req.Params[0]
		res = proto.NewResponse(proto.RES_TYPE_BULK)
		if v, ok := set[k]; ok {
			res.SetString(v)
		}
	} else {
		res = proto.NewErrorRes("wrong number of arguments for 'set' command")
	}
	return
}

// Set : Empty processor set
func (proc *SimpleProc) Set(req *proto.Request) (res *proto.Response, err error) {
	if len(req.Params) == 2 {
		k := req.Params[0]
		v := req.Params[1]
		set[k] = v
		res = proto.NewResponse(proto.RES_TYPE_STATE)
		res.SetState("OK")
	} else {
		res = proto.NewErrorRes("wrong number of arguments for 'set' command")
	}
	return
}

// Hset : Empty processor hset
func (proc *SimpleProc) Hset(req *proto.Request) (res *proto.Response, err error) {
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
	res.SetInt(strconv.Itoa(updateCount))
	return
}

// Hget : Empty processor hget
func (proc *SimpleProc) Hget(req *proto.Request) (res *proto.Response, err error) {
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

// Hgetall : Empty processor hgetall
func (proc *SimpleProc) Hgetall(req *proto.Request) (res *proto.Response, err error) {
	k := req.Params[0]
	var data map[string]string
	var ok bool
	if 1 == len(req.Params) {
		res = proto.NewResponse(proto.RES_TYPE_MULTI)
		if data, ok = hash[k]; ok {
			for field, value := range data {
				res.SetString(field)
				res.SetString(value)
			}
		}
	} else {
		res = proto.NewErrorRes("wrong number of arguments for 'set' command")
	}
	return
}

// Exec : Empty processor auth
func (proc *SimpleProc) Exec(req *proto.Request) (res []*proto.Response, err error) {
	if proc.isMulti {
		proc.isMulti = false
		for _, request := range proc.reqQue {
			cmd := strings.Title(request.Cmd)

			v := reflect.ValueOf(proc)
			method := v.MethodByName(cmd)
			result := method.Call([]reflect.Value{reflect.ValueOf(request)})
			if !result[0].IsNil() {
				res = append(res, result[0].Interface().(*proto.Response))
			} else {
				res = append(res, proto.NewErrorRes("Process `"+cmd+"` error"))
			}
		}
	} else {
		res = append(res, proto.NewErrorRes("EXEC without MULTI"))
	}
	return
}

func init() {
	set = make(map[string]string)
	hash = make(map[string](map[string]string))
}
