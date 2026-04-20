package application

import (
	"context"
	"errors"
	"fmt"
	"net"
	"notification_service/internal/config"
	"notification_service/internal/service"
	"notification_service/internal/storage/pg"
	apigrpc "notification_service/internal/transport/grpc"
	"notification_service/pkg/postgresql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rogue0026/marketplace-proto_v2/gen/notification_service/pb"
	"google.golang.org/grpc"
)

type App struct {
	DBPool     *pgxpool.Pool
	GRPCServer *grpc.Server
	Cfg        *config.AppConfig
}

func New(ctx context.Context, cfgPath string) (*App, error) {
	appCfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load service configuration: %w", err)
	}

	pool, err := postgresql.Pool(ctx, appCfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres pool: %w", err)
	}

	grpcServer := grpc.NewServer()

	notificationsRepository := pg.NewNotificationsRepository(pool)
	notificationsService := service.New(notificationsRepository)
	grpcHandler := apigrpc.NewHandler(notificationsService)

	pb.RegisterNotificationServiceServer(grpcServer, grpcHandler)

	return &App{
		DBPool:     pool,
		GRPCServer: grpcServer,
		Cfg:        appCfg,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", a.Cfg.GRPCServerAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", a.Cfg.GRPCServerAddress, err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- a.GRPCServer.Serve(l)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		a.Stop(shutdownCtx)
		return nil
	case err = <-serveErr:
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return fmt.Errorf("grpc server stopped with error: %w", err)
	}
}

func (a *App) Stop(ctx context.Context) {
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
