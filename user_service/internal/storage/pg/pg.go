package pg

import (
	"context"
	"errors"
	"fmt"
	"user_service/internal/apperrors"
	"user_service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

/*
CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	username VARCHAR(64) UNIQUE NOT NULL,
	password_hash VARCHAR(512) NOT NULL

);

CREATE TABLE IF NOT EXISTS wallets (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT REFERENCES users (id) ON DELETE CASCADE,
	balance BIGINT NOT NULL DEFAULT 0

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
	ON CONFLICT (username) DO NOTHING
	RETURNING id`

const CreateWalletSQL = `
	INSERT INTO wallets (user_id) 
	VALUES ($1)`

func (r *UsersRepository) CreateUser(ctx context.Context, username string, passwordHash string) (uint64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID uint64
	err = tx.QueryRow(ctx, CreateUserSQL, username, passwordHash).Scan(&userID)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("user with username=%s %w", username, apperrors.ErrAlreadyExists)
		}

		return 0, fmt.Errorf("failed to create user: %w", err)

	}

	if _, err = tx.Exec(ctx, CreateWalletSQL, userID); err != nil {
		return 0, fmt.Errorf("failed to create wallet: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to finish transaction: %w", err)
	}

	return userID, nil
}

const DeleteUserSQL = `
	DELETE FROM users
	WHERE id = $1`

func (r *UsersRepository) DeleteUser(ctx context.Context, userID uint64) error {
	ct, err := r.pool.Exec(ctx, DeleteUserSQL, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user=%d: %w", userID, err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("delete, user with id=%d: %w", userID, apperrors.ErrNotFound)
	}

	return nil
}

const AddMoneySQL = `
	UPDATE wallets
	SET balance = balance + $2
	WHERE user_id = $1
`

func (r *UsersRepository) AddMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	ct, err := r.pool.Exec(ctx, AddMoneySQL, userID, moneyAmount)
	if err != nil {
		return fmt.Errorf("failed to add money to user=%d: %w", userID, err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("add money, user=%d %w", userID, apperrors.ErrNotFound)
	}

	return nil
}

const WriteOffMoneySQL = `
	UPDATE wallets
	SET balance = balance - $2
	WHERE user_id = $1 AND balance >= $2
`

const UserExistsSQL = `SELECT EXISTS(SELECT id FROM users WHERE id = $1)`

func (r *UsersRepository) WriteOffMoney(ctx context.Context, userID uint64, moneyAmount uint64) error {
	ct, err := r.pool.Exec(ctx, WriteOffMoneySQL, userID, moneyAmount)
	if err != nil {
		return fmt.Errorf("failed to write off money to user=%d: %w", userID, err)
	}

	if ct.RowsAffected() == 0 {
		var userExists bool
		if err := r.pool.QueryRow(ctx, UserExistsSQL, userID).Scan(&userExists); err != nil {
			return fmt.Errorf("failed to check for user existence: %w", err)
		}

		if userExists {
			return fmt.Errorf("user=%d: %w", userID, apperrors.ErrNotEnoughMoney)
		}

		return fmt.Errorf("user=%d %w", userID, apperrors.ErrNotFound)
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
		return fmt.Errorf("failed to add product into basket. user=%d, product=%d: %w", userID, productID, err)
	}

	return nil
}

const DeleteProductFromBasketSQL = `
	DELETE FROM basket
	WHERE user_id = $1 AND product_id = $2`

func (r *UsersRepository) DeleteProductFromBasket(ctx context.Context, userID uint64, productID uint64) error {
	ct, err := r.pool.Exec(ctx, DeleteProductFromBasketSQL, userID, productID)
	if err != nil {
		return fmt.Errorf("failed to delete product=%d for user=%d from basket: %w", productID, userID, err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("delete product=%d for user=%d from basket: %w", productID, userID, apperrors.ErrNotFound)
	}

	return nil
}

const GetBasketSQL = `SELECT id, user_id, product_id, product_quantity FROM basket WHERE user_id = $1`

func (r *UsersRepository) GetBasket(ctx context.Context, userID uint64) ([]*domain.BasketItem, error) {
	rows, err := r.pool.Query(ctx, GetBasketSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user basket, user=%d: %w", userID, err)
	}
	defer rows.Close()

	basket := make([]*domain.BasketItem, 0)
	for rows.Next() {
		item := domain.BasketItem{}
		err = rows.Scan(&item.ID, &item.UserID, &item.ProductID, &item.ProductQuantity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row while fetching user basket, user=%d: %w", userID, err)
		}
		basket = append(basket, &item)
	}
	fmt.Println(basket)

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to get user basket: %w", err)
	}

	if len(basket) == 0 {
		return nil, fmt.Errorf("basket is empty: %w", apperrors.ErrNotFound)
	}

	return basket, nil
}

const ClearBasketSQL = `DELETE FROM basket WHERE user_id = $1`

func (r *UsersRepository) ClearBasket(ctx context.Context, userID uint64) error {
	_, err := r.pool.Exec(ctx, ClearBasketSQL, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user basket info: %w", err)
	}

	return nil
}
