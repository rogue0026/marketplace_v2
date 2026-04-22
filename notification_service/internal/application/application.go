package application

import (
	"context"
	"errors"
	"fmt"
	"net"
	"notification_service/internal/config"
	"notification_service/internal/messaging"
	"notification_service/internal/service"
	"notification_service/internal/storage/pg"
	apigrpc "notification_service/internal/transport/grpc"
	"notification_service/pkg/postgresql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rogue0026/marketplace-proto_v2/gen/notification_service/pb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type App struct {
	DBPool        *pgxpool.Pool
	GRPCServer    *grpc.Server
	Cfg           *config.AppConfig
	KafkaConsumer *messaging.Consumer
}

func New(ctx context.Context, cfgPath string) (*App, error) {
	appCfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load service configuration: %w", err)
	}
	if len(appCfg.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("failed to initialize kafka consumer: kafka-brokers is empty")
	}
	if appCfg.KafkaGroupID == "" {
		return nil, fmt.Errorf("failed to initialize kafka consumer: kafka-group-id is empty")
	}

	pool, err := postgresql.Pool(ctx, appCfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres pool: %w", err)
	}

	grpcServer := grpc.NewServer()

	notificationsRepository := pg.NewNotificationsRepository(pool)
	notificationsService := service.New(notificationsRepository)
	notificationsConsumer := messaging.NewConsumer(
		appCfg.KafkaBrokers,
		appCfg.KafkaGroupID,
		notificationsRepository,
	)
	grpcHandler := apigrpc.NewHandler(notificationsService)

	pb.RegisterNotificationServiceServer(grpcServer, grpcHandler)

	return &App{
		DBPool:        pool,
		GRPCServer:    grpcServer,
		Cfg:           appCfg,
		KafkaConsumer: notificationsConsumer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", a.Cfg.GRPCServerAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", a.Cfg.GRPCServerAddress, err)
	}

	g, runCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := a.GRPCServer.Serve(l)
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("grpc server stopped with error: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		return a.KafkaConsumer.Run(runCtx)
	})

	<-runCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	a.Stop(shutdownCtx)

	return g.Wait()
}

func (a *App) Stop(ctx context.Context) {
	if a.KafkaConsumer != nil {
		_ = a.KafkaConsumer.Close()
	}

	stopped := make(chan struct{})
	go func() {
		a.GRPCServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		a.GRPCServer.Stop()
	case <-stopped:
	}

	a.DBPool.Close()
}
