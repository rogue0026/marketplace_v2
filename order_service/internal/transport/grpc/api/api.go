package api

import (
	"context"
	"errors"
	"order_service/internal/apperrors"
	"order_service/internal/service"

	"github.com/rogue0026/marketplace-proto_v2/gen/order_service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	orderService *service.OrderService
	pb.UnimplementedOrderServiceServer
}

func New(orderService *service.OrderService) *Handler {
	h := &Handler{
		orderService: orderService,
	}

	return h
}

func (h *Handler) CreateOrder(ctx context.Context, in *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	orderID, err := h.orderService.CreateOrder(ctx, in.UserId)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.FailedPrecondition, "unable to create order, user basket is empty")
		}
		
		return nil, err
	}

	resp := &pb.CreateOrderResponse{OrderId: orderID}

	return resp, nil
}
