package api

import (
	"context"
	"product_service/internal/domain"
	"product_service/internal/service"

	"product_service/internal/transport/grpc/errmap"

	"github.com/rogue0026/marketplace-proto_v2/gen/product_service/pb"
	"google.golang.org/protobuf/types/known/emptypb"
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
		return nil, errmap.MapError(err)
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
		return nil, errmap.MapError(err)
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
		return nil, errmap.MapError(err)
	}

	resp := &pb.AddNewProductResponse{
		ProductId: productID,
	}

	return resp, nil
}

func (h *Handler) ReserveProducts(ctx context.Context, in *pb.ReserveProductsRequest) (*emptypb.Empty, error) {
	products := make([]*domain.Reservation, 0, len(in.Products))
	for _, elem := range in.Products {
		products = append(products, &domain.Reservation{
			ProductID: elem.ProductId,
			Quantity:  elem.Quantity,
		})
	}

	err := h.ProductService.ReserveProducts(ctx, in.OrderId, products)
	if err != nil {
		return nil, errmap.MapError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) CancelReservationsForOrder(ctx context.Context, in *pb.CancelReservationsForOrderRequest) (*emptypb.Empty, error) {
	err := h.ProductService.CancelReservationsForOrder(ctx, in.OrderId)
	if err != nil {
		return nil, errmap.MapError(err)
	}

	return &emptypb.Empty{}, nil
}
