package product_service

import (
	"context"
	"fmt"
	"order_service/internal/apperrors"
	"order_service/internal/domain"

	"github.com/rogue0026/marketplace-proto_v2/gen/product_service/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProductServiceClient struct {
	grpcClient pb.ProductServiceClient
}

func mapErr(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return err
	}

	var reason string
	for _, d := range s.Details() {
		errInfo, ok := d.(*errdetails.ErrorInfo)
		if ok {
			reason = errInfo.Reason
		}
	}

	switch {
	case s.Code() == codes.FailedPrecondition && reason == pb.Reason_NOT_ENOUGH_PRODUCTS.String():
		return apperrors.ErrNotEnoughProducts
	default:
		return err
	}
}

func NewProductServiceClient(ccAddress string) (*ProductServiceClient, error) {
	cc, err := grpc.NewClient(ccAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("client, product service: %w", err)
	}

	grpcClient := pb.NewProductServiceClient(cc)

	clientWrapper := &ProductServiceClient{
		grpcClient: grpcClient,
	}

	return clientWrapper, nil
}

func (c *ProductServiceClient) ReserveProducts(ctx context.Context, orderID uint64, products []*domain.Reservation) (*emptypb.Empty, error) {
	reservations := make([]*pb.ReserveProductsRequest_Reservation, 0)
	for _, p := range products {
		reservations = append(reservations, &pb.ReserveProductsRequest_Reservation{
			ProductId: p.ProductID,
			Quantity:  p.Quantity,
		})
	}

	resp, err := c.grpcClient.ReserveProducts(ctx, &pb.ReserveProductsRequest{
		OrderId:  orderID,
		Products: reservations,
	})

	if err != nil {
		return nil, fmt.Errorf("client, product service, reserve products: %w", mapErr(err))
	}

	return resp, nil
}

func (c *ProductServiceClient) CancelReservationForOrder(ctx context.Context, orderID uint64) (*emptypb.Empty, error) {
	resp, err := c.grpcClient.CancelReservationsForOrder(ctx, &pb.CancelReservationsForOrderRequest{
		OrderId: orderID,
	})

	if err != nil {
		return nil, fmt.Errorf("client, product service, cancel reservation: %w", mapErr(err))
	}

	return resp, nil
}
