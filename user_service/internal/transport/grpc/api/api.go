package api

import (
	"context"
	"errors"
	"user_service/internal/apperrors"
	"user_service/internal/service"

	"github.com/rogue0026/marketplace-proto_v2/gen/user_service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	UserService *service.UserService
	pb.UnimplementedUserServiceServer
}

func NewHandler(s *service.UserService) *Handler {
	return &Handler{
		UserService: s,
	}
}

func (h *Handler) CreateNewUser(ctx context.Context, in *pb.CreateNewUserRequest) (*pb.CreateNewUserResponse, error) {
	userID, err := h.UserService.CreateNewUser(ctx, in.Username, in.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.CreateNewUserResponse{
		UserId: userID,
	}

	return resp, nil
}

func (h *Handler) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	err := h.UserService.DeleteUser(ctx, in.UserId)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) AddProductToBasket(ctx context.Context, in *pb.AddProductToBasketRequest) (*emptypb.Empty, error) {
	err := h.UserService.AddProductToBasket(ctx, in.UserId, in.ProductId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) GetUserBasket(ctx context.Context, in *pb.GetUserBasketRequest) (*pb.GetUserBasketResponse, error) {
	userBasketInfo, err := h.UserService.GetBasket(ctx, in.UserId)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "basket is empty")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &pb.GetUserBasketResponse{
		Products: make([]*pb.GetUserBasketResponse_Product, 0),
	}

	for _, item := range userBasketInfo {
		resp.Products = append(
			resp.Products,
			&pb.GetUserBasketResponse_Product{
				Id:       item.ProductID,
				Name:     item.ProductName,
				Price:    item.ProductPrice,
				Quantity: item.ProductQuantityInBasket,
			})
	}

	return resp, nil
}

func (h *Handler) ClearBasket(ctx context.Context, in *pb.ClearBasketRequest) (*emptypb.Empty, error) {
	err := h.UserService.ClearBasket(ctx, in.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) AddMoney(ctx context.Context, in *pb.AddMoneyRequest) (*emptypb.Empty, error) {
	err := h.UserService.AddMoney(ctx, in.UserId, in.MoneyAmount)

	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) WriteOffMoney(ctx context.Context, in *pb.WriteOffMoneyRequest) (*emptypb.Empty, error) {
	err := h.UserService.WriteOffMoney(ctx, in.UserId, in.MoneyAmount)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}

		if errors.Is(err, apperrors.ErrNotEnoughMoney) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
