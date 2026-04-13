package user_service

import (
	"context"
	"fmt"
	"order_service/internal/domain"

	"github.com/rogue0026/marketplace-proto_v2/gen/user_service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserServiceClient struct {
	grpcClient pb.UserServiceClient
}

func NewUserServiceClient(ccAddress string) (*UserServiceClient, error) {
	cc, err := grpc.NewClient(ccAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection with user service: %w", err)
	}

	grpcClient := pb.NewUserServiceClient(cc)

	c := &UserServiceClient{
		grpcClient: grpcClient,
	}

	return c, nil
}

func (c *UserServiceClient) GetUserBasket(ctx context.Context, userID uint64) ([]*domain.Product, error) {
	resp, err := c.grpcClient.GetUserBasket(ctx, &pb.GetUserBasketRequest{UserId: userID})
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("failed to request data from user basket: %w", err)
		}

		if s.Code() == codes.NotFound {
			return nil, fmt.Errorf("basket is empty: %w", err)
		}
	}

	productsInBasket := make([]*domain.Product, 0, len(resp.Products))
	for _, elem := range resp.Products {
		productsInBasket = append(productsInBasket, &domain.Product{
			Id:       elem.Id,
			Name:     elem.Name,
			Price:    elem.Price,
			Quantity: elem.Quantity,
		})
	}

	return productsInBasket, nil
}

func (c *UserServiceClient) ClearUserBasket(ctx context.Context, userID uint64) error {
	_, err := c.grpcClient.ClearBasket(ctx, &pb.ClearBasketRequest{UserId: userID})
	if err != nil {
		return fmt.Errorf("failed to make request for clear user basket: %w", err)
	}

	return nil
}
