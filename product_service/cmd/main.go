package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"product_service/internal/application"
	"product_service/pkg/migrations"
	"syscall"
)

func main() {
	ctx := context.Background()
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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		app.Stop()
	}()

	fmt.Printf("starting product service at %s\n", app.Cfg.GRPCServerAddress)
	err = app.Run()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("service stopped")
}
