package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"user_service/internal/application"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := application.New(ctx, "./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		cancel()
		app.Stop()
	}()

	fmt.Printf("starting user service at %s\n", app.Cfg.GRPCServerAddress)
	err = app.Run(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("service stopped")
}
