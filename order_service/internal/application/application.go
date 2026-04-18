package application

import (
	"context"
	"fmt"
	"net"
	ps "order_service/internal/clients/product_service"
	us "order_service/internal/clients/user_service"
	"order_service/internal/config"
	"order_service/internal/messaging"
	"order_service/internal/service"
	"order_service/internal/storage/pg"
	"order_service/internal/transport/grpc/api"
	"order_service/pkg/postgresql"

	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/rogue0026/marketplace-proto_v2/gen/order_service/pb"
	"google.golang.org/grpc"
)

type App struct {
	OutboxRelay *messaging.Relay
	Cfg         *config.Config
	GRPCServer  *grpc.Server
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

	// creating outbox relay
	topics := []contracts.Topic{
		contracts.OrderEvents,
	}
	outboxRelay := messaging.NewRelay(pool, appCfg.KafkaBrokers, topics)

	// creating repositories
	ordersRepository := pg.NewOrdersRepository(pool)

	// creating clients
	userServiceClient, err := us.NewUserServiceClient(appCfg.UserServiceAddress)
	if err != nil {
		return nil, err
	}

	productServiceClient, err := ps.NewProductServiceClient(appCfg.ProductServiceAddress)
	if err != nil {
		return nil, err
	}

	// creating service
	OrderService := service.NewOrderService(ordersRepository, userServiceClient, productServiceClient)

	// creating grpc server
	grpcServer := grpc.NewServer()
	grpcHandler := api.New(OrderService)

	pb.RegisterOrderServiceServer(grpcServer, grpcHandler)

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

	go func() {
		fmt.Printf("starting outbox relay\n")
		a.OutboxRelay.Run(ctx)
	}()

	fmt.Printf("user service: starting grpc server at %s\n", a.Cfg.GRPCServerAddress)
	err = a.GRPCServer.Serve(l)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
	a.GRPCServer.GracefulStop()
}
