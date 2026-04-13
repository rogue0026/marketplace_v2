package application

import (
	"context"
	"fmt"
	us "order_service/internal/clients/user_service"
	"order_service/internal/config"
	"order_service/internal/service"
	"order_service/internal/storage/pg"
	"order_service/internal/transport/grpc/api"
	"order_service/pkg/postgresql"

	"github.com/rogue0026/marketplace-proto_v2/gen/order_service/pb"
	"google.golang.org/grpc"
)

type App struct {
	Cfg        *config.Config
	GRPCServer *grpc.Server
}

func New(ctx context.Context, configPath string) (*App, error) {

	// loading service configuration
	appCfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load service configuration: %w", err)
	}

	// creating database pool
	pool, err := postgresql.Pool(ctx, appCfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// creating repositories
	ordersRepository := pg.NewOrdersRepository(pool)

	// creating clients
	userServiceClient, err := us.NewUserServiceClient(appCfg.UserServiceAddress)
	if err != nil {
		return nil, err
	}

	// creating service
	OrderService := service.NewOrderService(ordersRepository, userServiceClient)

	// creating grpc server
	grpcServer := grpc.NewServer()
	grpcHandler := api.New(OrderService)

	pb.RegisterOrderServiceServer(grpcServer, grpcHandler)

	// creating application instance
	a := &App{
		Cfg:        appCfg,
		GRPCServer: grpcServer,
	}

	return a, nil
}
