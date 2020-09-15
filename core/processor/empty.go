package processor

import (
	"gredissimulate/core/proto"
)

// Processor : Processor interface
type Processor interface {
	Get(*proto.Request) (*proto.Response, error)
	Set(*proto.Request) (*proto.Response, error)
	Ping(*proto.Request) (*proto.Response, error)
	HSet(*proto.Request) (*proto.Response, error)
	HGet(*proto.Request) (*proto.Response, error)
	Auth(*proto.Request) (*proto.Response, error)
}

// ProcessReq : Process request
func ProcessReq(proc Processor, req *proto.Request) (res *proto.Response, err error) {
	switch req.Cmd {
	case "get":
		res, err = proc.Get(req)
	case "set":
		res, err = proc.Set(req)
	case "ping":
		res, err = proc.Ping(req)
	case "hset":
		res, err = proc.HSet(req)
	case "hget":
		res, err = proc.HGet(req)
	case "auth":
		res, err = proc.Auth(req)
	}
	return
}

// EmptyProc : Do nothing
type EmptyProc struct{}

// Get : Empty processor get
func (proc *EmptyProc) Get(req *proto.Request) (res *proto.Response, err error) {
	return
}

// Set : Empty processor set
func (proc *EmptyProc) Set(req *proto.Request) (res *proto.Response, err error) {
	return
}

// Ping : Empty processor ping
func (proc *EmptyProc) Ping(req *proto.Request) (res *proto.Response, err error) {
	return
}

// HSet : Empty processor hset
func (proc *EmptyProc) HSet(req *proto.Request) (res *proto.Response, err error) {
	return
}

// HGet : Empty processor hget
func (proc *EmptyProc) HGet(req *proto.Request) (res *proto.Response, err error) {
	return
}

// Auth : Empty processor auth
func (proc *EmptyProc) Auth(req *proto.Request) (res *proto.Response, err error) {
	return
}
