package api

import (
	"context"
	"errors"
	"product_service/internal/apperrors"
	"product_service/internal/domain"
	"product_service/internal/service"

	"github.com/rogue0026/marketplace-proto_v2/gen/product_service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	ProductService *service.ProductService
	pb.UnimplementedProductServiceServer
}

func NewHandler(s *service.ProductService) *Handler {
	h := &Handler{
		ProductService: s,
	}

	return h
}

func (h *Handler) ProductCatalog(ctx context.Context, in *pb.ProductCatalogRequest) (*pb.ProductCatalogResponse, error) {
	products, err := h.ProductService.ProductsPaginated(ctx, in.Page, in.Size)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "products not found")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	pbProducts := make([]*pb.ProductCatalogResponse_Product, 0, len(products))
	for _, elem := range products {
		pbProducts = append(pbProducts, &pb.ProductCatalogResponse_Product{
			Id:       elem.ID,
			Name:     elem.Name,
			Price:    elem.Price,
			Quantity: elem.Quantity,
		})
	}

	resp := &pb.ProductCatalogResponse{
		Products: pbProducts,
	}

	return resp, nil
}

func (h *Handler) ProductsById(ctx context.Context, in *pb.ProductsByIdRequest) (*pb.ProductsByIdResponse, error) {
	products, err := h.ProductService.ProductsByID(ctx, in.IdList)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "no data")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.ProductsByIdResponse{
		Products: make([]*pb.ProductsByIdResponse_Product, 0, len(products)),
	}

	for _, elem := range products {
		resp.Products = append(
			resp.Products,
			&pb.ProductsByIdResponse_Product{
				Id:       elem.ID,
				Name:     elem.Name,
				Price:    elem.Price,
				Quantity: elem.Quantity,
			},
		)
	}

	return resp, nil
}

func (h *Handler) AddNewProduct(ctx context.Context, in *pb.AddNewProductRequest) (*pb.AddNewProductResponse, error) {
	productID, err := h.ProductService.AddNewProduct(ctx, &domain.Product{
		Name:     in.Name,
		Price:    in.Price,
		Quantity: in.Quantity,
	})

	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, "invalid input data")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.AddNewProductResponse{
		ProductId: productID,
	}

	return resp, nil
}

//func (h *Handler) DeleteProduct(ctx context.Context, in *pb.DeleteProductRequest) (*emptypb.Empty, error) {}
