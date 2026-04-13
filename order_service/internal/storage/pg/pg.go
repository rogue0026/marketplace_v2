package pg

import (
	"context"
	"fmt"
	"order_service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrdersRepository struct {
	pool *pgxpool.Pool
}

func NewOrdersRepository(pool *pgxpool.Pool) *OrdersRepository {
	repo := &OrdersRepository{
		pool: pool,
	}

	return repo
}

/*
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    order_status VARCHAR(128),
);

CREATE TABLE IF NOT EXISTS orders_content (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES orders (id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    product_quantity BIGINT NOT NULL,
    product_price_per_unit BIGINT NOT NULL
);
*/

const CreateOrderSQL = `
	INSERT INTO orders (user_id, total_price, order_status)
	VALUES ($1, $2, $3) RETURNING id`

const AddOrderContentSQL = `
	INSERT INTO orders_content (order_id, user_id, product_id, product_quantity, product_price_per_unit)
	VALUES ($1, $2, $3, $4, $5)`

func (r *OrdersRepository) CreateOrder(ctx context.Context, userID uint64, orderItems []*domain.OrderItem) (uint64, error) {
	// compute order total price
	var OrderTotalPrice uint64

	for _, elem := range orderItems {
		OrderTotalPrice += elem.ProductQuantity * elem.ProductPricePerUnit
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction while creating order: %w", err)
	}
	defer tx.Rollback(ctx)

	var orderID uint64
	err = tx.QueryRow(
		ctx,
		CreateOrderSQL,
		userID,
		OrderTotalPrice,
		domain.StatusWaitingForPayment,
	).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	for _, elem := range orderItems {
		_, err := tx.Exec(
			ctx,
			AddOrderContentSQL,
			orderID,
			userID,
			elem.ProductID,
			elem.ProductQuantity,
			elem.ProductPricePerUnit,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to insert row: order=%d user=%d: %w", orderID, userID, err)
		}

	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to finish transaction while creating order: %w", err)
	}

	return orderID, nil
}
