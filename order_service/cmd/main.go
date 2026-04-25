package main

import (
	"context"
	"fmt"
	"order_service/internal/application"
	"order_service/pkg/migrations"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := application.New(ctx, "./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = migrations.Run("migrations", fmt.Sprintf("%s?sslmode=disable", app.Cfg.DatabaseURL))
	if err != nil {
		fmt.Printf("applying migrations: %s\n", err.Error())
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		cancel()
		app.Stop()
	}()

	fmt.Printf("starting order service at %s\n", app.Cfg.GRPCServerAddress)
	err = app.Run(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("service stopped")
}
