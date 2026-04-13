package main

import (
	"context"
	"fmt"
	"net"
	"order_service/internal/application"
)

func main() {
	ctx := context.Background()

	app, err := application.New(ctx, "./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	l, err := net.Listen("tcp", app.Cfg.GRPCServerAddress)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("starting order service at %s\n", app.Cfg.GRPCServerAddress)
	err = app.GRPCServer.Serve(l)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}
