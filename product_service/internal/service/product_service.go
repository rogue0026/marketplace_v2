package service

import (
	"context"
	"product_service/internal/domain"
)

type ProductsRepository interface {
	GetProductsPaginated(ctx context.Context, page uint64, size uint64) ([]*domain.Product, error)
	GetProductsByID(ctx context.Context, IDList []uint64) ([]*domain.Product, error)
	NewProduct(ctx context.Context, name string, price uint64, quantity uint64) (uint64, error)
	DeleteProduct(ctx context.Context, productID uint64) error
}
type ProductService struct {
	products ProductsRepository
}

func New(repo ProductsRepository) *ProductService {
	s := &ProductService{
		products: repo,
	}

	return s
}

func (s *ProductService) ProductsPaginated(ctx context.Context, page uint64, size uint64) ([]*domain.Product, error) {
	products, err := s.products.GetProductsPaginated(ctx, page, size)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) ProductsByID(ctx context.Context, IDList []uint64) ([]*domain.Product, error) {
	products, err := s.products.GetProductsByID(ctx, IDList)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) AddNewProduct(ctx context.Context, p *domain.Product) (uint64, error) {
	productID, err := s.products.NewProduct(ctx, p.Name, p.Price, p.Quantity)
	if err != nil {
		return 0, err
	}

	return productID, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, productID uint64) error {
	err := s.products.DeleteProduct(ctx, productID)
	if err != nil {
		return err
	}

	return nil
}
