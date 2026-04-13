package application

import (
	"context"
	"fmt"
	"github.com/rogue0026/marketplace-proto_v2/gen/user_service/pb"
	"google.golang.org/grpc"
	ps "user_service/internal/clients/product_service"
	"user_service/internal/config"
	"user_service/internal/service"
	"user_service/internal/storage/pg"
	"user_service/internal/transport/grpc/api"
	"user_service/pkg/postgresql"
)

type App struct {
	Cfg        *config.AppConfig
	GRPCServer *grpc.Server
}

func New(ctx context.Context, configPath string) (*App, error) {
	// loading service configuration
	appCfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load service configuration: %w", err)
	}

	// creating connection pool
	pool, err := postgresql.Pool(ctx, appCfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// creating repository
	usersRepository := pg.NewUsersRepository(pool)

	// creating clients
	productServiceClient, err := ps.NewProductServiceClient(appCfg.ProductServiceAddress)
	if err != nil {
		return nil, err
	}

	// creating service
	userService := service.New(usersRepository, productServiceClient)

	// creating grpc server
	grpcServer := grpc.NewServer()
	h := api.NewHandler(userService)
	pb.RegisterUserServiceServer(grpcServer, h)

	// creating application instance
	a := &App{
		Cfg:        appCfg,
		GRPCServer: grpcServer,
	}

	return a, nil
}
