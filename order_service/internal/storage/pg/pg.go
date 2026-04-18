package pg

import (
	"context"
	"errors"
	"fmt"
	"order_service/internal/apperrors"
	"order_service/internal/domain"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/rogue0026/kafka-contracts/events"
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

const WriteToOutboxSQL = `
	INSERT INTO outbox (topic_name, partition_key, payload, published_at)
	VALUES ($1, $2, $3, $4)
`

/*
CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id),
    user_id BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    topic_name TEXT NOT NULL,
    partition_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(64) NOT NULL DEFAULT 'PENDING',
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    published_at TIMESTAMP,
    CONSTRAINT status_one_of CHECK ( status in ('PENDING', 'PROCESSED') )
);

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
	var OrderTotalPrice uint64

	for _, elem := range orderItems {
		OrderTotalPrice += elem.ProductQuantity * elem.ProductPricePerUnit
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("repo, create order: %w", err)
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
		return 0, fmt.Errorf("repo, create order, scan row: %w", err)
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
			return 0, fmt.Errorf("repo, create order, add order content, order_id=%d: %w", orderID, err)
		}
	}

	e := events.OrderCreated{
		OrderID: orderID,
	}

	msg, err := events.NewMessage(e, contracts.OrderService, time.Now())
	if err != nil {
		return 0, fmt.Errorf("repo, create order, create message: %w", err)
	}

	rawMsg, err := msg.Raw()
	if err != nil {
		return 0, fmt.Errorf("repo, create order, create raw message: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		WriteToOutboxSQL,
		contracts.OrderEvents,
		strconv.FormatUint(orderID, 10),
		rawMsg,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("repo, create order, writing to outbox: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("repo, create order: %w", err)
	}

	return orderID, nil
}

const OrderContentInfoSQL = `
	SELECT product_id, product_quantity, product_price_per_unit
	FROM orders_content
	WHERE order_id = $1
	`

func (r *OrdersRepository) OrderContentInfo(ctx context.Context, orderID uint64) ([]*domain.OrderItem, error) {
	rows, err := r.pool.Query(
		ctx,
		OrderContentInfoSQL,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("repo, order content info: %w", err)
	}
	defer rows.Close()

	orderContent := make([]*domain.OrderItem, 0)
	for rows.Next() {
		item := domain.OrderItem{}
		err = rows.Scan(&item.ProductID, &item.ProductQuantity, &item.ProductPricePerUnit)
		if err != nil {
			return nil, fmt.Errorf("repo, order content, scan row: %w", err)
		}
		orderContent = append(orderContent, &item)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("repo, order content info: %w", err)
	}

	return orderContent, nil
}

const OrderGeneralInfoSQL = `SELECT user_id, total_price FROM orders WHERE id = $1`

func (r *OrdersRepository) OrderGeneralInfo(ctx context.Context, orderID uint64) (map[string]uint64, error) {
	var userID uint64
	var orderTotalPrice uint64

	err := r.pool.QueryRow(ctx, OrderGeneralInfoSQL, orderID).Scan(&userID, &orderTotalPrice)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repo, order general info: %w", apperrors.ErrOrderNotFound)
		}

		return nil, fmt.Errorf("repo, order general info: %w", err)
	}

	m := make(map[string]uint64)
	m["user_id"] = userID
	m["total_price"] = orderTotalPrice

	return m, nil
}

const ChangeOrderStatusSQL = `
	UPDATE orders 
	SET order_status = $2
	WHERE id = $1
`

func (r *OrdersRepository) ChangeOrderStatus(ctx context.Context, orderID uint64, statusValue string) error {
	_, err := r.pool.Exec(ctx, ChangeOrderStatusSQL, orderID, statusValue)
	if err != nil {
		return fmt.Errorf("failed to change order status (order_id=%d): %w", orderID, err)
	}

	return nil
}

const CreatePaymentSQL = `
	INSERT INTO payments (order_id, user_id, total_price) 
	VALUES ($1, $2, $3)
	RETURNING id
`

func (r *OrdersRepository) CreatePayment(ctx context.Context, orderID uint64, userID uint64, totalPrice uint64) (uint64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("repo, create payment: %w", err)
	}
	defer tx.Rollback(ctx)

	var paymentID uint64
	err = tx.QueryRow(
		ctx,
		CreatePaymentSQL,
		orderID,
		userID,
		totalPrice,
	).Scan(&paymentID)
	if err != nil {
		return 0, fmt.Errorf("repo, create payment: %w", err)
	}

	e := events.OrderPayedFor{
		PaymentID: paymentID,
	}

	msg, err := events.NewMessage(e, contracts.OrderService, time.Now())
	if err != nil {
		return 0, fmt.Errorf("repo, create payment, create message: %w", err)
	}

	rawMsg, err := msg.Raw()
	if err != nil {
		return 0, fmt.Errorf("repo, create payment, create raw message: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		WriteToOutboxSQL,
		contracts.OrderEvents,
		strconv.FormatUint(paymentID, 10),
		rawMsg,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("repo, create payment, writing to outbox: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("repo, create payment: %w", err)
	}

	return paymentID, nil
}
