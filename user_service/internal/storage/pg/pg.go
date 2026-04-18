package pg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
	"user_service/internal/apperrors"
	"user_service/internal/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rogue0026/kafka-contracts/contracts"
	"github.com/rogue0026/kafka-contracts/events"
)

type UsersRepository struct {
	pool *pgxpool.Pool
}

func NewUsersRepository(connPool *pgxpool.Pool) *UsersRepository {
	r := &UsersRepository{
		pool: connPool,
	}

	return r
}

const WriteToOutboxSQL = `
	INSERT INTO outbox (topic_name, partition_key, payload, published_at)
	VALUES ($1, $2, $3, $4)
`

/*
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

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password_hash VARCHAR(512) NOT NULL,
    CONSTRAINT username_unique UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS wallets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT balance_is_positive CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS basket (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users (id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    product_quantity BIGINT NOT NULL CHECK (product_quantity > 0) DEFAULT 1,
    CONSTRAINT user_id_product_id_unique UNIQUE (user_id, product_id)
);
*/

const CreateUserSQL = `
	INSERT INTO users (username, password_hash)
	VALUES ($1, $2)
	RETURNING id`

const CreateWalletSQL = `
	INSERT INTO wallets (user_id) 
	VALUES ($1)`

func (r *UsersRepository) CreateUser(ctx context.Context, username string, passwordHash string) (uint64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("repo, create user: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID uint64
	err = tx.QueryRow(ctx, CreateUserSQL, username, passwordHash).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgerrcode.UniqueViolation == pgErr.Code &&
			pgErr.ConstraintName == "username_unique" {
			return 0, fmt.Errorf("repo, create user: %w", apperrors.ErrUsernameAlreadyTaken)
		}

		return 0, fmt.Errorf("repo, create user: %w", err)
	}

	if _, err = tx.Exec(ctx, CreateWalletSQL, userID); err != nil {
		return 0, fmt.Errorf("repo, create user: %w", err)
	}

	event := &events.UserCreated{
		UserID: userID,
	}
	msg, err := events.NewMessage(event, contracts.UserService, time.Now())
	if err != nil {
		return 0, fmt.Errorf("repo, create user, make message: %w", err)
	}

	rawMsg, err := msg.Raw()
	if err != nil {
		return 0, fmt.Errorf("repo, create user, make raw message: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		WriteToOutboxSQL,
		contracts.UserEvents,
		strconv.FormatUint(userID, 10),
		rawMsg,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("repo, create user, writing to outbox: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("repo, create user: %w", err)
	}

	return userID, nil
}

const DeleteUserSQL = `
	DELETE FROM users
	WHERE id = $1`

func (r *UsersRepository) DeleteUser(ctx context.Context, userID uint64) error {
	ct, err := r.pool.Exec(ctx, DeleteUserSQL, userID)
	if err != nil {
		return fmt.Errorf("repo, delete user, user_id=%d: %w", userID, err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("repo, delete user, user_id=%d: %w", userID, apperrors.ErrUserNotFound)
	}

	return nil
}

const AddMoneySQL = `
	UPDATE wallets
	SET balance = balance + $2
	WHERE user_id = $1
`

func (r *UsersRepository) AddMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repo, add money: %w", err)
	}
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx, AddMoneySQL, userID, moneyAmount)
	if err != nil {
		return fmt.Errorf("repo, add money: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("repo, add money, user_id=%d %w", userID, apperrors.ErrUserNotFound)
	}

	event := &events.FundsAdded{
		UserID: userID,
	}
	msg, err := events.NewMessage(event, contracts.UserService, time.Now())
	if err != nil {
		return fmt.Errorf("repo, add money, user_id=%d: %w", userID, err)
	}

	rawMsg, err := msg.Raw()
	if err != nil {
		return fmt.Errorf("repo, add money, user_id=%d: %w", userID, err)
	}

	_, err = tx.Exec(
		ctx,
		WriteToOutboxSQL,
		contracts.WalletEvents,
		strconv.FormatUint(userID, 10),
		rawMsg,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("repo, add money, user_id=%d: %w", userID, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("repo, add money: %w", err)
	}

	return nil
}

const WriteOffMoneySQL = `
	UPDATE wallets
	SET balance = balance - $2
	WHERE user_id = $1
`

func (r *UsersRepository) WriteOffMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repo, write off money, user_id=%d: %w", userID, err)
	}
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx, WriteOffMoneySQL, userID, moneyAmount)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == pgerrcode.CheckViolation &&
			pgErr.ConstraintName == "balance_is_positive" {
			return fmt.Errorf("repo, write off money, user_id=%d: %w", userID, apperrors.ErrNotEnoughMoney)
		}

		return fmt.Errorf("repo, write off money: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("repo, write off money: %w", apperrors.ErrUserNotFound)
	}

	event := events.FundsDebitted{
		UserID: userID,
	}

	msg, err := events.NewMessage(event, contracts.UserService, time.Now())
	if err != nil {
		return fmt.Errorf("repo, write off money: %w", err)
	}

	rawMsg, err := msg.Raw()
	if err != nil {
		return fmt.Errorf("repo, write off money: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		WriteToOutboxSQL,
		contracts.WalletEvents,
		strconv.FormatUint(userID, 10),
		rawMsg,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("repo, write off money: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("repo, write off money: %w", err)
	}

	return nil
}

const AddProductToBasketSQL = `
	INSERT INTO basket (user_id, product_id)
	VALUES ($1, $2)
	ON CONFLICT ON CONSTRAINT user_id_product_id_unique
	DO UPDATE SET product_quantity = basket.product_quantity + excluded.product_quantity;

`

func (r *UsersRepository) AddProductToBasket(ctx context.Context, userID uint64, productID uint64) error {
	_, err := r.pool.Exec(ctx, AddProductToBasketSQL, userID, productID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return fmt.Errorf("repo, add product to basket, user_id=%d: %w", userID, apperrors.ErrUserNotFound)
		}

		return fmt.Errorf("repo, add product to basket: %w", err)
	}

	return nil
}

const DeleteProductFromBasketSQL = `
	DELETE FROM basket
	WHERE user_id = $1 AND product_id = $2`

func (r *UsersRepository) DeleteProductFromBasket(ctx context.Context, userID uint64, productID uint64) error {
	_, err := r.pool.Exec(ctx, DeleteProductFromBasketSQL, userID, productID)
	if err != nil {
		return fmt.Errorf("repo, delete product from basket: %w", err)
	}

	return nil
}

const GetBasketSQL = `SELECT product_id, product_quantity FROM basket WHERE user_id = $1`

func (r *UsersRepository) GetBasket(ctx context.Context, userID uint64) ([]*domain.BasketItem, error) {
	rows, err := r.pool.Query(ctx, GetBasketSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("repo, get basket: %w", err)
	}
	defer rows.Close()

	basket := make([]*domain.BasketItem, 0)
	for rows.Next() {
		item := domain.BasketItem{}
		err = rows.Scan(&item.ProductID, &item.ProductQuantity)
		if err != nil {
			return nil, fmt.Errorf("repo, get basket, scan row: %w", err)
		}

		basket = append(basket, &item)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("repo, get basket: %w", err)
	}

	if len(basket) == 0 {
		return nil, fmt.Errorf("repo, get basket: %w", apperrors.ErrEmptyBasket)
	}

	return basket, nil
}

const ClearBasketSQL = `DELETE FROM basket WHERE user_id = $1`

func (r *UsersRepository) ClearBasket(ctx context.Context, userID uint64) error {
	_, err := r.pool.Exec(ctx, ClearBasketSQL, userID)
	if err != nil {
		return fmt.Errorf("repo, clear basket: %w", err)
	}

	return nil
}
