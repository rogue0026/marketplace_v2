package main

import (
	"context"
	"fmt"
	"net"
	"product_service/internal/application"
)

func main() {
	ctx := context.Background()
	app, err := application.New(ctx, "./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}

	lis, err := net.Listen("tcp", app.Cfg.GRPCServerAddress)
	if err != nil {
		fmt.Println("error while open tcp connection", err.Error())
		return
	}
	fmt.Println("starting product service at", app.Cfg.GRPCServerAddress)
	err = app.GRPCServer.Serve(lis)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
