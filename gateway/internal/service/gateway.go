package service

import (
	"context"
	os "gateway/internal/clients/order_service"
	ps "gateway/internal/clients/product_service"
	us "gateway/internal/clients/user_service"
	"gateway/internal/domain"
)

type Gateway struct {
	ProductService *ps.ProductService
	UserService    *us.UserService
	OrderService   *os.OrderService
}

func New(productService *ps.ProductService, userService *us.UserService, orderService *os.OrderService) *Gateway {
	return &Gateway{
		ProductService: productService,
		UserService:    userService,
		OrderService:   orderService,
	}
}

func (g *Gateway) CreateNewUser(ctx context.Context, username string, password string) (uint64, error) {
	userID, err := g.UserService.CreateNewUser(ctx, username, password)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (g *Gateway) DeleteUser(ctx context.Context, userID uint64) error {
	err := g.UserService.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gateway) ProductCatalogPaginated(ctx context.Context, page uint64, size uint64) ([]*domain.Product, error) {
	products, err := g.ProductService.ProductCatalog(ctx, page, size)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (g *Gateway) ProductsByIDList(ctx context.Context, idList []uint64) ([]*domain.Product, error) {
	products, err := g.ProductService.ProductsByIDList(ctx, idList)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (g *Gateway) AddProductToBasket(ctx context.Context, userID uint64, productID uint64) error {
	err := g.UserService.AddProductToBasket(ctx, userID, productID)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gateway) AddMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	err := g.UserService.AddMoney(ctx, userID, moneyAmount)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gateway) CreateOrder(ctx context.Context, userID uint64) (uint64, error) {
	orderID, err := g.OrderService.CreateOrder(ctx, userID)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func (g *Gateway) PayForOrder(ctx context.Context, orderID uint64) (uint64, error) {
	paymentID, err := g.OrderService.PayForOrder(ctx, orderID)
	if err != nil {
		return 0, err
	}

	return paymentID, nil
}
