package pg

import (
	"context"
	"errors"
	"fmt"
	"product_service/internal/apperrors"
	"product_service/internal/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductsRepository struct {
	pool *pgxpool.Pool
}

func NewProductsRepository(connPool *pgxpool.Pool) *ProductsRepository {
	r := &ProductsRepository{
		pool: connPool,
	}

	return r
}

const GetProductsPaginatedSQL = `
	SELECT id, name, price, quantity FROM products ORDER BY id LIMIT $1 OFFSET $2`

func (r *ProductsRepository) GetProductsPaginated(ctx context.Context, page uint64, size uint64) ([]*domain.Product, error) {
	if page == 0 {
		page = 1
	}
	offset := (page - 1) * size

	rows, err := r.pool.Query(ctx, GetProductsPaginatedSQL, size, offset)
	if err != nil {
		return nil, fmt.Errorf("products paginated. failed to fetch rows: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		p := &domain.Product{}
		err = rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity)
		if err != nil {
			return nil, fmt.Errorf("products paginated. error while scanning row: %w", err)
		}
		products = append(products, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("products paginated. error: %w", err)
	}

	if len(products) == 0 {
		return nil, fmt.Errorf("products paginated. error: %w", apperrors.ErrNotFound)
	}

	return products, nil
}

const ProductsByIDSQL = `SELECT id, name, price, quantity FROM products WHERE id = ANY($1)`

func (r *ProductsRepository) GetProductsByID(ctx context.Context, IDList []uint64) ([]*domain.Product, error) {
	rows, err := r.pool.Query(ctx, ProductsByIDSQL, IDList)
	if err != nil {
		return nil, fmt.Errorf("products by id. failed to fetch rows: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		p := &domain.Product{}
		err = rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity)
		if err != nil {
			return nil, fmt.Errorf("products by id. failed to scan row: %w", err)
		}

		products = append(products, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("products by id. error: %w", err)
	}

	if len(products) == 0 {
		return nil, fmt.Errorf("products by id. error: %w", apperrors.ErrNotFound)
	}

	return products, nil
}

const NewProductSQL = `INSERT INTO products (name, price, quantity) VALUES ($1, $2, $3) RETURNING id`

func (r *ProductsRepository) NewProduct(ctx context.Context, name string, price uint64, quantity uint64) (uint64, error) {
	var productID uint64

	err := r.pool.QueryRow(ctx, NewProductSQL, name, price, quantity).Scan(&productID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.CheckViolation {
			return 0, fmt.Errorf(
				"failed to create new product. name=%s price=%d quantity=%d: %w",
				name,
				price,
				quantity,
				apperrors.ErrInvalidArgument)
		}

		return 0, fmt.Errorf("error while creating new product: %w", err)
	}

	return productID, nil
}

const DeleteProductSQL = `DELETE FROM products WHERE id = $1`

func (r *ProductsRepository) DeleteProduct(ctx context.Context, productID uint64) error {
	ct, err := r.pool.Exec(ctx, DeleteProductSQL, productID)
	if err != nil {
		return fmt.Errorf("failed to delete product=%d: %w", productID, err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete product=%d: %w", productID, apperrors.ErrNotFound)
	}

	return nil
}
