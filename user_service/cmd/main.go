package main

import (
	"context"
	"fmt"
	"net"
	"user_service/internal/application"
)

//func main() {
//	ctx := context.Background()
//	app, err := application.New(ctx, "./config.yaml")
//	if err != nil {
//		fmt.Println(err.Error())
//		return
//	}
//
//	lis, err := net.Listen("tcp", app.Cfg.GRPCServerAddress)
//	go func() {
//		fmt.Println("starting user service at", app.Cfg.GRPCServerAddress)
//		err = app.GRPCServer.Serve(lis)
//	}()
//
//	cc, err := grpc.NewClient("localhost:4001", grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		fmt.Println(err.Error())
//		return
//	}
//	testClient := pb.NewUserServiceClient(cc)
//	_, err = testClient.WriteOffMoney(ctx, &pb.WriteOffMoneyRequest{
//		UserId:      12,
//		MoneyAmount: 1000000,
//	})
//	if err != nil {
//		fmt.Println(err.Error())
//	}
//
//	stop := make(chan os.Signal, 1)
//	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
//	fmt.Println("waiting to stop signal")
//	<-stop
//}

func main() {
	ctx := context.Background()
	app, err := application.New(ctx, "./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	lis, err := net.Listen("tcp", app.Cfg.GRPCServerAddress)
	fmt.Println("starting user service at", app.Cfg.GRPCServerAddress)
	err = app.GRPCServer.Serve(lis)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
