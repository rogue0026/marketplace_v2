package api

import (
	"context"
	"order_service/internal/service"
	"order_service/internal/transport/grpc/errmap"

	"github.com/rogue0026/marketplace-proto_v2/gen/order_service/pb"
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
		return nil, errmap.MapError(err)
	}

	resp := &pb.CreateOrderResponse{OrderId: orderID}

	return resp, nil
}

func (h *Handler) PayForOrder(ctx context.Context, in *pb.PayForOrderRequest) (*pb.PayForOrderResponse, error) {
	paymentID, err := h.orderService.PayForOrder(ctx, in.OrderId)
	if err != nil {
		return nil, errmap.MapError(err)
	}

	resp := &pb.PayForOrderResponse{PaymentId: paymentID}

	return resp, nil
}
