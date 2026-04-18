package order_service

import (
	"context"
	"errors"
	"fmt"
	"gateway/internal/apperrors"

	"github.com/rogue0026/marketplace-proto_v2/gen/order_service/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type errKey struct {
	code   codes.Code
	reason string
}

var errorsMap = map[errKey]error{
	{
		codes.FailedPrecondition,
		pb.Reason_NOT_ENOUGH_PRODUCTS.String(),
	}: apperrors.ErrNotEnoughProducts,

	{
		codes.FailedPrecondition,
		pb.Reason_NOT_ENOUGH_MONEY.String(),
	}: apperrors.ErrNotEnoughMoney,

	{
		codes.FailedPrecondition,
		pb.Reason_BASKET_IS_EMPTY.String(),
	}: apperrors.ErrBasketIsEmpty,

	{
		codes.Internal,
		pb.Reason_PRODUCT_DOES_NOT_EXISTS.String(),
	}: apperrors.ErrProductsNotFound,

	{
		codes.NotFound,
		pb.Reason_USER_NOT_FOUND.String(),
	}: apperrors.ErrUserNotFound,

	{
		codes.NotFound,
		pb.Reason_ORDER_NOT_FOUND.String(),
	}: apperrors.ErrOrderNotFound,
}

var ErrEmptyReason = errors.New("reason is empty")

func extractReason(s *status.Status) (string, error) {
	for _, d := range s.Details() {
		errInfo, ok := d.(*errdetails.ErrorInfo)
		if ok {
			return errInfo.Reason, nil
		}
	}

	return "", ErrEmptyReason
}

func mapErr(err error) error {
	st, ok := status.FromError(err)
	if ok {
		reason, err := extractReason(st)
		if err != nil {
			return err
		}

		k := errKey{
			code:   st.Code(),
			reason: reason,
		}

		appErr, ok := errorsMap[k]
		if ok {
			return appErr
		}
	}

	return err
}

type OrderService struct {
	client pb.OrderServiceClient
}

func NewOrderService(ccAddr string) (*OrderService, error) {
	cc, err := grpc.NewClient(ccAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("client, order service: %w", err)
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

		return 0, fmt.Errorf("client, order service, create order: %w", mapErr(err))
	}

	return resp.OrderId, nil
}

func (c *OrderService) PayForOrder(ctx context.Context, orderID uint64) (uint64, error) {
	resp, err := c.client.PayForOrder(ctx, &pb.PayForOrderRequest{OrderId: orderID})
	if err != nil {
		return 0, fmt.Errorf("client, order service, create order: %w", mapErr(err))
	}

	return resp.PaymentId, nil
}
