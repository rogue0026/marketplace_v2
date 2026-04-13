package service

import (
	"context"
	"fmt"
	us "order_service/internal/clients/user_service"
	"order_service/internal/domain"
)

type OrdersRepository interface {
	CreateOrder(ctx context.Context, userID uint64, orderItems []*domain.OrderItem) (uint64, error)
}

type OrderService struct {
	userServiceClient *us.UserServiceClient
	repo              OrdersRepository
}

func NewOrderService(repo OrdersRepository, userServiceClient *us.UserServiceClient) *OrderService {
	s := &OrderService{
		userServiceClient: userServiceClient,
		repo:              repo,
	}

	return s
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uint64) (uint64, error) {
	productsInBasket, err := s.userServiceClient.GetUserBasket(ctx, userID)
	if err != nil {
		fmt.Printf("requesting data from user service: %s", err.Error())
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
