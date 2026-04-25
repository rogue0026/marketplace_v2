package main

import (
	"context"
	"fmt"
	"notification_service/internal/application"
	"notification_service/pkg/migrations"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app, err := application.New(ctx, "./config.yaml")
	if err != nil {
		fmt.Printf("creating app instance: %s\n", err.Error())
		return
	}

	err = migrations.Run("migrations", fmt.Sprintf("%s?sslmode=disable", app.Cfg.DatabaseURL))
	if err != nil {
		fmt.Printf("applying migrations: %s\n", err.Error())
		return
	}

	fmt.Printf("starting notification service at %s\n", app.Cfg.GRPCServerAddress)
	err = app.Run(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
