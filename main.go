package main

import (
	"context"
	"runtime"
	"sync"

	"gredissimulate/config"
	"gredissimulate/core"
	"gredissimulate/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	logConf := logger.LogConf{LogPath: "", CacheSize: 100}
	log, _ := logger.NewServer(logConf)

	server, err := core.NewServer(core.ServerConf{Port: config.GetListenPort()})
	if nil != err {
		panic(err)
	}

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		log.Start(ctx)
		wg.Done()
	}()

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
