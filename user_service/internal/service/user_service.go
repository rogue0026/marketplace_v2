package service

import (
	"context"
	"fmt"
	ps "user_service/internal/clients/product_service"
	"user_service/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, username string, passwordHash string) (uint64, error)
	DeleteUser(ctx context.Context, userID uint64) error
	AddMoney(ctx context.Context, userID uint64, moneyAmount uint64) error
	WriteOffMoney(ctx context.Context, userID uint64, moneyAmount uint64) error
	AddProductToBasket(ctx context.Context, userID uint64, productID uint64) error
	DeleteProductFromBasket(ctx context.Context, userID uint64, productID uint64) error
	GetBasket(ctx context.Context, userID uint64) ([]*domain.BasketItem, error)
	ClearBasket(ctx context.Context, userID uint64) error
}

type UserService struct {
	productsClient *ps.ProductService
	users          UsersRepository
}

func New(repo UsersRepository, productsClient *ps.ProductService) *UserService {
	s := &UserService{
		productsClient: productsClient,
		users:          repo,
	}

	return s
}

func (s *UserService) CreateNewUser(ctx context.Context, username string, password string) (uint64, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("service, create user: %w", err)
	}

	userID, err := s.users.CreateUser(ctx, username, string(hashed))
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uint64) error {
	err := s.users.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) AddMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	err := s.users.AddMoney(ctx, userID, moneyAmount)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) WriteOffMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	err := s.users.WriteOffMoney(ctx, userID, moneyAmount)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) AddProductToBasket(ctx context.Context, userID uint64, productID uint64) error {
	err := s.users.AddProductToBasket(ctx, userID, productID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteProductFromBasket(ctx context.Context, userID uint64, productID uint64) error {
	err := s.users.DeleteProductFromBasket(ctx, userID, productID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetBasket(ctx context.Context, userID uint64) (map[uint64]*domain.BasketItemAggregated, error) {
	basket, err := s.users.GetBasket(ctx, userID)
	if err != nil {
		return nil, err
	}

	basketInfoAggregated := make(map[uint64]*domain.BasketItemAggregated)
	idList := make([]uint64, 0)

	for _, elem := range basket {
		basketInfoAggregated[elem.ProductID] = &domain.BasketItemAggregated{
			ProductID:               elem.ProductID,
			ProductQuantityInBasket: elem.ProductQuantity,
		}
		idList = append(idList, elem.ProductID)
	}

	productsInfo, err := s.productsClient.ProductsByIDList(ctx, idList)
	if err != nil {
		return nil, err
	}

	for _, elem := range productsInfo {
		itemAggregated, ok := basketInfoAggregated[elem.ID]
		if ok {
			itemAggregated.ProductName = elem.Name
			itemAggregated.ProductPrice = elem.Price
		}
	}

	return basketInfoAggregated, nil
}

func (s *UserService) ClearBasket(ctx context.Context, userID uint64) error {
	err := s.users.ClearBasket(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
