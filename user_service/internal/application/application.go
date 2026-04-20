package application

import (
	"context"
	"fmt"
	"net"
	ps "user_service/internal/clients/product_service"
	"user_service/internal/config"
	"user_service/internal/messaging"
	"user_service/internal/service"
	"user_service/internal/storage/pg"
	"user_service/internal/transport/grpc/api"
	"user_service/pkg/postgresql"

	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/rogue0026/marketplace-proto_v2/gen/user_service/pb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type App struct {
	OutboxRelay *messaging.Relay
	Cfg         *config.AppConfig
	GRPCServer  *grpc.Server
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

	// creating outbox relay
	topics := []contracts.Topic{
		contracts.WalletEvents,
		contracts.UserEvents,
	}
	outboxRelay := messaging.NewRelay(
		pool,
		appCfg.KafkaBrokers,
		topics,
	)

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
		OutboxRelay: outboxRelay,
		Cfg:         appCfg,
		GRPCServer:  grpcServer,
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", a.Cfg.GRPCServerAddress)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := a.GRPCServer.Serve(l)
		if err != nil {
			return err
		}

		return nil
	})

	a.OutboxRelay.Run(ctx)

	return g.Wait()
}

func (a *App) Stop() {
	a.GRPCServer.GracefulStop()
}
