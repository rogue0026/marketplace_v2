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

func mapErr(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch {
	case s.Code() == codes.AlreadyExists:
		return apperrors.ErrUserAlreadyExists
	case s.Code() == codes.NotFound:
		return apperrors.ErrUserNotFound
	case s.Code() == codes.FailedPrecondition:
		return apperrors.ErrNotEnoughMoney
	default:
		return err
	}
}

type UserService struct {
	client pb.UserServiceClient
}

func NewUserService(ccString string) (*UserService, error) {
	cc, err := grpc.NewClient(ccString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("client, user service: %w", err)
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
		return 0, fmt.Errorf("client, user service, create user: %w", mapErr(err))
	}

	return resp.UserId, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := s.client.DeleteUser(ctx, &pb.DeleteUserRequest{UserId: userID})
	if err != nil {
		return fmt.Errorf("client, user service, delete user: %w", mapErr(err))
	}

	return nil
}

func (s *UserService) AddMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	_, err := s.client.AddMoney(ctx, &pb.AddMoneyRequest{
		UserId:      userID,
		MoneyAmount: moneyAmount,
	})

	if err != nil {
		return fmt.Errorf("client, user service, add money: %w", mapErr(err))
	}

	return nil
}

func (s *UserService) AddProductToBasket(ctx context.Context, userID uint64, productID uint64) error {
	_, err := s.client.AddProductToBasket(ctx, &pb.AddProductToBasketRequest{
		UserId:    userID,
		ProductId: productID,
	})

	if err != nil {
		return fmt.Errorf("client, user service, add product to basket: %w", mapErr(err))
	}

	return nil
}
