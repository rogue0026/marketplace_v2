package api

import (
	"context"
	"user_service/internal/service"
	"user_service/internal/transport/grpc/errmap"

	"github.com/rogue0026/marketplace-proto_v2/gen/user_service/pb"
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
		return nil, errmap.MapError(err)
	}

	resp := &pb.CreateNewUserResponse{
		UserId: userID,
	}

	return resp, nil
}

func (h *Handler) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	err := h.UserService.DeleteUser(ctx, in.UserId)
	if err != nil {
		return nil, errmap.MapError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) AddProductToBasket(ctx context.Context, in *pb.AddProductToBasketRequest) (*emptypb.Empty, error) {
	err := h.UserService.AddProductToBasket(ctx, in.UserId, in.ProductId)
	if err != nil {
		return nil, errmap.MapError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) GetUserBasket(ctx context.Context, in *pb.GetUserBasketRequest) (*pb.GetUserBasketResponse, error) {
	userBasketInfo, err := h.UserService.GetBasket(ctx, in.UserId)
	if err != nil {
		return nil, errmap.MapError(err)
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
		return nil, errmap.MapError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) AddMoney(ctx context.Context, in *pb.AddMoneyRequest) (*emptypb.Empty, error) {
	err := h.UserService.AddMoney(ctx, in.UserId, in.MoneyAmount)

	if err != nil {
		return nil, errmap.MapError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) WriteOffMoney(ctx context.Context, in *pb.WriteOffMoneyRequest) (*emptypb.Empty, error) {
	err := h.UserService.WriteOffMoney(ctx, in.UserId, in.MoneyAmount)
	if err != nil {
		return nil, errmap.MapError(err)
	}

	return &emptypb.Empty{}, nil
}
