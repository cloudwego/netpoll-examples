package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/netpoll"
)

func main() {
	eventLoop, err := netpoll.NewEventLoop(
		handle,
		netpoll.WithOnPrepare(prepare),
		netpoll.WithOnConnect(connect),
		netpoll.WithReadTimeout(2*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener, err := netpoll.CreateListener("tcp", ":9997")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := eventLoop.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	eventLoop.Shutdown(ctx)
}
