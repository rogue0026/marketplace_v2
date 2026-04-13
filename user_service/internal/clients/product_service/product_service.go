package product_service

import (
	"context"
	"fmt"
	"user_service/internal/apperrors"
	"user_service/internal/domain"

	"github.com/rogue0026/marketplace-proto_v2/gen/product_service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type ProductService struct {
	grpcClient pb.ProductServiceClient
}

func NewProductServiceClient(ccAddr string) (*ProductService, error) {
	cc, err := grpc.NewClient(ccAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection with product service: %w", err)
	}

	grpcClient := pb.NewProductServiceClient(cc)

	clientWrapper := &ProductService{
		grpcClient: grpcClient,
	}

	return clientWrapper, nil
}

func (c *ProductService) ProductsByIDList(ctx context.Context, IDList []uint64) ([]*domain.Product, error) {
	resp, err := c.grpcClient.ProductsById(ctx, &pb.ProductsByIdRequest{IdList: IDList})
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("failed to request data from product service: %w", err)
		}
		if s.Code() == codes.NotFound {
			return nil, fmt.Errorf("products %w", apperrors.ErrNotFound)
		}
	}

	productsInfo := make([]*domain.Product, 0)
	for _, elem := range resp.Products {
		p := &domain.Product{
			ID:       elem.Id,
			Name:     elem.Name,
			Price:    elem.Price,
			Quantity: elem.Quantity,
		}
		productsInfo = append(productsInfo, p)
	}

	return productsInfo, nil
}
