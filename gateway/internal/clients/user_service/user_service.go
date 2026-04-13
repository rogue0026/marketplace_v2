package user_service

import (
	"context"
	"fmt"
	"gateway/internal/apperrors"

	"github.com/rogue0026/marketplace-proto_v2/gen/user_service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserService struct {
	client pb.UserServiceClient
}

func NewUserService(ccString string) (*UserService, error) {
	cc, err := grpc.NewClient(ccString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection with user service: %w", err)
	}

	client := pb.NewUserServiceClient(cc)
	s := &UserService{
		client: client,
	}

	return s, nil
}

func (s *UserService) CreateNewUser(ctx context.Context, username string, password string) (uint64, error) {
	resp, err := s.client.CreateNewUser(ctx, &pb.CreateNewUserRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return 0, fmt.Errorf("failed to create new user: %w", err)
		}

		if s.Code() == codes.AlreadyExists {
			return 0, fmt.Errorf("error: user %w", apperrors.ErrAlreadyExists)
		}
	}

	return resp.UserId, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := s.client.DeleteUser(ctx, &pb.DeleteUserRequest{UserId: userID})
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return fmt.Errorf("failed to request delete user=%d: %w", userID, err)
		}
		if s.Code() == codes.NotFound {
			return fmt.Errorf("the requested user=%d %w", userID, apperrors.ErrNotFound)
		}
	}

	return nil
}

func (s *UserService) AddMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	_, err := s.client.AddMoney(ctx, &pb.AddMoneyRequest{
		UserId:      userID,
		MoneyAmount: moneyAmount,
	})

	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return fmt.Errorf("failed to request add money to user=%d: %w", userID, err)
		}
		if s.Code() == codes.NotFound {
			return fmt.Errorf("user %w", apperrors.ErrNotFound)
		}
	}

	return nil
}

func (s *UserService) AddProductToBasket(ctx context.Context, userID uint64, productID uint64) error {
	_, err := s.client.AddProductToBasket(ctx, &pb.AddProductToBasketRequest{
		UserId:    userID,
		ProductId: productID,
	})

	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return fmt.Errorf("failed to request to add product into basket: %w", err)
		}

		return s.Err()
	}

	return nil
}
