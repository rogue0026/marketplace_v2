package main

import (
	"context"
	"fmt"
	"notification_service/internal/application"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app, err := application.New(ctx, "./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("starting notification service at %s\n", app.Cfg.GRPCServerAddress)
	err = app.Run(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
