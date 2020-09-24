package main

import (
	"context"
	"runtime"
	"sync"

	"gredissimulate/config"
	"gredissimulate/core"
	"gredissimulate/core/processor"
	"gredissimulate/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	logConf := logger.LogConf{LogPath: "", CacheSize: 100}
	log, _ := logger.NewServer(logConf)

	// Create a new server with Simple command processor
	server, err := core.NewServer(core.ServerConf{Port: config.GetListenPort(), Passwd: config.GetPasswd()}, processor.NewSimpleProc)
	if nil != err {
		panic(err)
	}

	var wg sync.WaitGroup

	// Start logger service
	go func() {
		wg.Add(1)
		log.Start(ctx)
		wg.Done()
	}()

	// Start redis service
	go func() {
		wg.Add(1)
		server.Start(ctx)
		wg.Done()
	}()

	runtime.Gosched()

	// cancel()
	defer cancel()
	wg.Wait()

}
