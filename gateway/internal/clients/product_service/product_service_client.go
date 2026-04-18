package product_service

import (
	"context"
	"fmt"
	"gateway/internal/apperrors"
	"gateway/internal/domain"

	"github.com/rogue0026/marketplace-proto_v2/gen/product_service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type ProductService struct {
	client pb.ProductServiceClient
}

func mapErr(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch {
	case s.Code() == codes.FailedPrecondition:
		return apperrors.ErrNotEnoughProducts

	case s.Code() == codes.InvalidArgument:
		return apperrors.ErrInvalidUserInput

	case s.Code() == codes.NotFound:
		return apperrors.ErrProductsNotFound

	default:
		return err
	}
}

func NewProductService(ccString string) (*ProductService, error) {
	cc, err := grpc.NewClient(ccString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("client, product service: %w", err)
	}

	client := pb.NewProductServiceClient(cc)

	s := &ProductService{
		client: client,
	}

	return s, nil
}

func (s *ProductService) ProductCatalog(ctx context.Context, page uint64, size uint64) ([]*domain.Product, error) {
	in := &pb.ProductCatalogRequest{
		Page: page,
		Size: size,
	}

	resp, err := s.client.ProductCatalog(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("client, product service, product catalog: %w", mapErr(err))
	}

	products := make([]*domain.Product, 0, len(resp.Products))
	for _, p := range resp.Products {
		product := &domain.Product{
			ID:       p.Id,
			Name:     p.Name,
			Price:    p.Price,
			Quantity: p.Quantity,
		}
		products = append(products, product)
	}

	return products, nil
}

func (s *ProductService) ProductsByIDList(ctx context.Context, idList []uint64) ([]*domain.Product, error) {
	in := &pb.ProductsByIdRequest{
		IdList: idList,
	}

	resp, err := s.client.ProductsById(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("client, product service, products by ids: %w", mapErr(err))
	}

	products := make([]*domain.Product, 0, len(resp.Products))
	for _, elem := range resp.Products {
		products = append(products, &domain.Product{
			ID:       elem.Id,
			Name:     elem.Name,
			Price:    elem.Price,
			Quantity: elem.Quantity,
		})
	}

	return products, nil
}
