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
		return nil, fmt.Errorf("repository, products paginated, failed to fetch rows: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		p := &domain.Product{}
		err = rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity)
		if err != nil {
			return nil, fmt.Errorf("repository, products paginated, failed to scan row: %w", err)
		}
		products = append(products, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("repository, products paginated: %w", err)
	}

	if len(products) == 0 {
		return nil, fmt.Errorf("repository, products paginated: %w", apperrors.ErrProductNotFound)
	}

	return products, nil
}

const ProductsByIDSQL = `SELECT id, name, price, quantity FROM products WHERE id = ANY($1)`

func (r *ProductsRepository) GetProductsByID(ctx context.Context, IDList []uint64) ([]*domain.Product, error) {
	rows, err := r.pool.Query(ctx, ProductsByIDSQL, IDList)
	if err != nil {
		return nil, fmt.Errorf("repo, products by id, failed to fetch rows: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		p := &domain.Product{}
		err = rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity)
		if err != nil {
			return nil, fmt.Errorf("repo, products by id. failed to scan row: %w", err)
		}

		products = append(products, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("repo, products by id: %w", err)
	}

	if len(products) == 0 {
		return nil, fmt.Errorf("repo, products by id: %w", apperrors.ErrProductNotFound)
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
			fmt.Printf("constraint name: %s\n", pgErr.ConstraintName) // todo: for debug
			switch {
			case pgErr.ConstraintName == "name_not_empty":
				return 0, fmt.Errorf("create product, %w: product name", apperrors.ErrInvalidUserInput)
			case pgErr.ConstraintName == "price_greater_zero":
				return 0, fmt.Errorf("create product, %w: product price", apperrors.ErrInvalidUserInput)
			case pgErr.ConstraintName == "quantity_greater_zero":
				return 0, fmt.Errorf("create product, %w: product quantity", apperrors.ErrInvalidUserInput)
			}
		}

		return 0, fmt.Errorf("create product: %w", err)
	}

	return productID, nil
}

const DeleteProductSQL = `DELETE FROM products WHERE id = $1`

func (r *ProductsRepository) DeleteProduct(ctx context.Context, productID uint64) error {
	ct, err := r.pool.Exec(ctx, DeleteProductSQL, productID)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("delete product: %w", apperrors.ErrProductNotFound)
	}

	return nil
}

const ReserveProductsSQL = `
	INSERT INTO reservations (order_id, product_id, quantity)
	VALUES ($1, $2, $3)`

const ReduceProductsSQL = `
	UPDATE products
	SET quantity = quantity - $2
	WHERE id = $1
`

func (r *ProductsRepository) ReserveProducts(ctx context.Context, orderID uint64, products []*domain.Reservation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repo, reserve products: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, p := range products {
		_, err = tx.Exec(
			ctx,
			ReserveProductsSQL,
			orderID,
			p.ProductID,
			p.Quantity,
		)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
				return fmt.Errorf(
					"repo, reserve products, product_id=%d: %w",
					p.ProductID,
					apperrors.ErrProductDoesNotExist,
				)
			}

			return fmt.Errorf("repo, reserve products: %w", err)
		}

		_, err = tx.Exec(
			ctx,
			ReduceProductsSQL,
			p.ProductID,
			p.Quantity,
		)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) &&
				pgErr.Code == pgerrcode.CheckViolation &&
				pgErr.ConstraintName == "quantity_greater_zero" {

				return fmt.Errorf(
					"repo, reserve products, product_id=%d: %w",
					p.ProductID,
					apperrors.ErrNotEnoughProducts,
				)

			}

			return fmt.Errorf(
				"repo, reserve products, product_id=%d: %w",
				p.ProductID,
				err,
			)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("repo, reserve products: %w", err)
	}

	return nil
}

const DeleteReservationSQL = `
	DELETE FROM reservations
	WHERE order_id = $1`

const ReservationInfoSQL = `
	SELECT product_id, quantity
	FROM reservations
	WHERE order_id = $1`

const IncreaseProductsSQL = `
	UPDATE products
	SET quantity = quantity + $2
	WHERE id = $1`

func (r *ProductsRepository) CancelReservation(ctx context.Context, orderID uint64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repo, cancel reservation: %w", err)
	}

	rows, err := tx.Query(ctx, ReservationInfoSQL, orderID)
	if err != nil {
		return fmt.Errorf("repo, cancel reservation: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.Reservation, 0)
	for rows.Next() {
		item := domain.Reservation{}
		err = rows.Scan(&item.ProductID, &item.Quantity)
		if err != nil {
			return fmt.Errorf("repo, cancel reservation, scanning row: %w", err)
		}

		items = append(items, &item)
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("repo, cancel reservation: %w", err)
	}

	for _, item := range items {
		_, err = tx.Exec(
			ctx,
			IncreaseProductsSQL,
			item.ProductID,
			item.Quantity,
		)
		if err != nil {
			return fmt.Errorf("repo, cancel reservation: %w", err)
		}

	}

	_, err = tx.Exec(ctx, DeleteReservationSQL, orderID)
	if err != nil {
		return fmt.Errorf("repo, cancel reservation: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("repo, cancel reservation: %w", err)
	}

	return nil
}
