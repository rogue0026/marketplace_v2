package service

import (
	"context"
	ps "order_service/internal/clients/product_service"
	us "order_service/internal/clients/user_service"
	"order_service/internal/domain"
)

type OrdersRepository interface {
	OrderContentInfo(ctx context.Context, orderID uint64) ([]*domain.OrderItem, error)
	CreateOrder(ctx context.Context, userID uint64, orderItems []*domain.OrderItem) (uint64, error)
	OrderGeneralInfo(ctx context.Context, orderID uint64) (map[string]uint64, error)
	ChangeOrderStatus(ctx context.Context, orderID uint64, statusValue string) error
	CreatePayment(ctx context.Context, orderID uint64, userID uint64, totalPrice uint64) (uint64, error)
}

type OrderService struct {
	userServiceClient    *us.UserServiceClient
	productServiceClient *ps.ProductServiceClient
	repo                 OrdersRepository
}

func NewOrderService(
	repo OrdersRepository,
	userServiceClient *us.UserServiceClient,
	productServiceClient *ps.ProductServiceClient,
) *OrderService {
	s := &OrderService{
		userServiceClient:    userServiceClient,
		productServiceClient: productServiceClient,
		repo:                 repo,
	}

	return s
}

func (s *OrderService) OrderContent(ctx context.Context, orderID uint64) ([]*domain.OrderItem, error) {
	orderContent, err := s.repo.OrderContentInfo(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return orderContent, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uint64) (uint64, error) {
	productsInBasket, err := s.userServiceClient.GetUserBasket(ctx, userID)
	if err != nil {
		return 0, err
	}

	orderItems := make([]*domain.OrderItem, 0, len(productsInBasket))
	for _, elem := range productsInBasket {
		orderItems = append(orderItems, &domain.OrderItem{
			ProductID:           elem.Id,
			ProductQuantity:     elem.Quantity,
			ProductPricePerUnit: elem.Price,
		})
	}

	orderID, err := s.repo.CreateOrder(ctx, userID, orderItems)
	if err != nil {
		return 0, err
	}

	err = s.userServiceClient.ClearUserBasket(ctx, userID)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func (s *OrderService) PayForOrder(ctx context.Context, orderID uint64) (uint64, error) {
	orderGeneralInfo, err := s.repo.OrderGeneralInfo(ctx, orderID)
	if err != nil {
		return 0, err
	}

	userID := orderGeneralInfo["user_id"]
	orderTotalPrice := orderGeneralInfo["total_price"]

	content, err := s.OrderContent(ctx, orderID)
	if err != nil {
		return 0, err
	}

	reservations := make([]*domain.Reservation, 0, len(content))
	for _, item := range content {
		reservations = append(reservations, &domain.Reservation{
			ProductID: item.ProductID,
			Quantity:  item.ProductQuantity,
		})
	}

	_, err = s.productServiceClient.ReserveProducts(ctx, orderID, reservations)
	if err != nil {
		return 0, err
	}

	err = s.userServiceClient.WriteOffMoney(ctx, userID, orderTotalPrice)
	if err != nil {
		_, err = s.productServiceClient.CancelReservationForOrder(ctx, orderID)
		if err != nil {
			return 0, err
		}

		return 0, err
	}

	err = s.repo.ChangeOrderStatus(ctx, orderID, domain.StatusPayedSuccessfully)
	if err != nil {
		return 0, err
	}

	paymentID, err := s.repo.CreatePayment(ctx, orderID, userID, orderTotalPrice)
	if err != nil {
		return 0, err
	}

	return paymentID, nil
}
