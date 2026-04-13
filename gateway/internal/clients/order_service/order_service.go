package order_service

import (
	"context"
	"fmt"
	"gateway/internal/apperrors"

	"github.com/rogue0026/marketplace-proto_v2/gen/order_service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type OrderService struct {
	client pb.OrderServiceClient
}

func NewOrderService(ccAddr string) (*OrderService, error) {
	cc, err := grpc.NewClient(ccAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection with order service: %w", err)
	}

	grpcClient := pb.NewOrderServiceClient(cc)
	s := &OrderService{
		client: grpcClient,
	}

	return s, nil
}

func (c *OrderService) CreateOrder(ctx context.Context, userID uint64) (uint64, error) {
	resp, err := c.client.CreateOrder(ctx, &pb.CreateOrderRequest{
		UserId: userID,
	})
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return 0, err
		}

		if s.Code() == codes.NotFound {
			return 0, fmt.Errorf("basket is empty: %w", apperrors.ErrNotFound)
		} else {
			return 0, fmt.Errorf("unknown error: %w", err)
		}
	}

	return resp.OrderId, nil
}
