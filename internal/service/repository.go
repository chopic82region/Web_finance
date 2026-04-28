package service

import (
	"context"
	"database/sql"
	"errors"
	"finance_tracker/internal/apperrors"
	"finance_tracker/internal/model"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *sql.DB
}

// UserRepository implements interfaces.User
type UserRepository struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if user == nil {
		return apperrors.ErrCreateUser
	}
	if err := IsValidEmail(user.Email); err != nil {
		return err
	}

	query := `
        INSERT INTO users (email, password_hash, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.ErrDuplicateEmail
		}
		return apperrors.ErrCreateUser
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	if user == nil {
		return apperrors.ErrUserNotFound
	}
	if user.ID == 0 {
		return apperrors.ErrUserNotFound
	}
	if err := IsValidEmail(user.Email); err != nil {
		return err
	}

	query := `
        UPDATE users
        SET email = $1, password_hash = $2, name = $3, updated_at = $4
        WHERE id = $5`

	user.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.ErrDuplicateEmail
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, user *model.User) error {
	if user == nil {
		return apperrors.ErrUserNotFound
	}
	if user.ID == 0 {
		return apperrors.ErrUserNotFound
	}

	query := `
        DELETE FROM users
        WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, user.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 = unique_violation
		return pgErr.Code == "23505"
	}
	return false
}

// AccountRepository implements interfaces.Account
type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepo(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *model.Account) error {
	if account == nil {
		return errors.New("account is nil")
	}

	query := `
		INSERT INTO accounts (user_id, name, type, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	return r.db.QueryRowContext(ctx, query,
		account.UserID,
		account.Name,
		account.Type,
		account.Balance,
		now,
		now,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
}

func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*model.Account, error) {
	query := `
		SELECT id, user_id, name, type, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1`

	var a model.Account
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID,
		&a.UserID,
		&a.Name,
		&a.Type,
		&a.Balance,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrAccountNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID int64) ([]model.Account, error) {
	query := `
		SELECT id, user_id, name, type, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Account
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Name,
			&a.Type,
			&a.Balance,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AccountRepository) Update(ctx context.Context, account *model.Account) error {
	if account == nil || account.ID == 0 {
		return apperrors.ErrAccountNotFound
	}

	query := `
		UPDATE accounts
		SET name = $1, type = $2, balance = $3, updated_at = $4
		WHERE id = $5`

	account.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, query,
		account.Name,
		account.Type,
		account.Balance,
		account.UpdatedAt,
		account.ID,
	)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return apperrors.ErrAccountNotFound
	}
	return nil
}

func (r *AccountRepository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return apperrors.ErrAccountNotFound
	}
	query := `DELETE FROM accounts WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return apperrors.ErrAccountNotFound
	}
	return nil
}

// TransactionRepository implements interfaces.Transactioons
type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepo(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *model.Transaction) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	query := `
		INSERT INTO transactions (user_id, account_id, category_id, type, amount, date, description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	return r.db.QueryRowContext(ctx, query,
		tx.UserID,
		tx.AccountID,
		tx.CategoryID,
		tx.Type,
		tx.Amount,
		tx.Date,
		tx.Description,
		now,
		now,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)
}

func (r *TransactionRepository) GetByID(ctx context.Context, id int64) (*model.Transaction, error) {
	query := `
		SELECT id, user_id, account_id, category_id, type, amount, date, description, created_at, updated_at
		FROM transactions
		WHERE id = $1`

	var t model.Transaction
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.UserID,
		&t.AccountID,
		&t.CategoryID,
		&t.Type,
		&t.Amount,
		&t.Date,
		&t.Description,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrTransactionNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TransactionRepository) GetByUserID(ctx context.Context, userID int64) ([]model.Transaction, error) {
	query := `
		SELECT id, user_id, account_id, category_id, type, amount, date, description, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY date DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.AccountID,
			&t.CategoryID,
			&t.Type,
			&t.Amount,
			&t.Date,
			&t.Description,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TransactionRepository) GetByAccountID(ctx context.Context, accountID int64) ([]model.Transaction, error) {
	query := `
		SELECT id, user_id, account_id, category_id, type, amount, date, description, created_at, updated_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY date DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.AccountID,
			&t.CategoryID,
			&t.Type,
			&t.Amount,
			&t.Date,
			&t.Description,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TransactionRepository) GetByDateRange(ctx context.Context, userID int64, from, to time.Time) ([]model.Transaction, error) {
	query := `
		SELECT id, user_id, account_id, category_id, type, amount, date, description, created_at, updated_at
		FROM transactions
		WHERE user_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.AccountID,
			&t.CategoryID,
			&t.Type,
			&t.Amount,
			&t.Date,
			&t.Description,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx *model.Transaction) error {
	if tx == nil || tx.ID == 0 {
		return apperrors.ErrTransactionNotFound
	}

	query := `
		UPDATE transactions
		SET account_id = $1, category_id = $2, type = $3, amount = $4, date = $5, description = $6, updated_at = $7
		WHERE id = $8`

	tx.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, query,
		tx.AccountID,
		tx.CategoryID,
		tx.Type,
		tx.Amount,
		tx.Date,
		tx.Description,
		tx.UpdatedAt,
		tx.ID,
	)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return apperrors.ErrTransactionNotFound
	}
	return nil
}

func (r *TransactionRepository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return apperrors.ErrTransactionNotFound
	}
	query := `DELETE FROM transactions WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return apperrors.ErrTransactionNotFound
	}
	return nil
}
