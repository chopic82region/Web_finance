package interfaces

import (
	"context"
	"finance_tracker/internal/model"
	"time"
)

type User interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, user *model.User) error
}

type Account interface {
	Create(ctx context.Context, account *model.Account) error
	GetByID(ctx context.Context, id int64) (*model.Account, error)
	GetByUserID(ctx context.Context, userID int64) ([]model.Account, error)
	Update(ctx context.Context, account *model.Account) error
    Delete(ctx context.Context, id int64) error
}

type Transactioons interface {
	Create(ctx context.Context, transaction *model.Transaction) error
    GetByID(ctx context.Context, id int64) (*model.Transaction, error)
    GetByUserID(ctx context.Context, userID int64) ([]model.Transaction, error)
    GetByAccountID(ctx context.Context, accountID int64) ([]model.Transaction, error)
    GetByDateRange(ctx context.Context, userID int64, from, to time.Time) ([]model.Transaction, error)
    Update(ctx context.Context, transaction *model.Transaction) error
    Delete(ctx context.Context, id int64) error
}
