package application

import (
	"context"
	"fmt"
	"net"
	"product_service/internal/config"
	"product_service/internal/service"
	"product_service/internal/storage/pg"
	"product_service/internal/transport/grpc/api"
	"product_service/pkg/postgresql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rogue0026/marketplace-proto_v2/gen/product_service/pb"
	"google.golang.org/grpc"
)

type App struct {
	connPool   *pgxpool.Pool
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
	productsRepo := pg.NewProductsRepository(pool)

	// creating service
	s := service.New(productsRepo)

	// creating grpc server
	grpcServer := grpc.NewServer()
	h := api.NewHandler(s)

	pb.RegisterProductServiceServer(grpcServer, h)

	// creating application instance
	app := &App{
		connPool:   pool,
		Cfg:        appCfg,
		GRPCServer: grpcServer,
	}

	return app, nil
}

func (a *App) Run() error {
	defer a.connPool.Close()

	l, err := net.Listen("tcp", a.Cfg.GRPCServerAddress)
	if err != nil {
		return nil
	}

	err = a.GRPCServer.Serve(l)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
	a.GRPCServer.GracefulStop()
}
