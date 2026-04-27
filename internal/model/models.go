package model

import "time"

// model/base.go – встраиваемая базовая модель
type BaseModel struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Account struct {
	BaseModel
	UserID  int64   `json:"user_id"  db:"user_id"`
	Name    string  `json:"name"     db:"name"`
	Type    string  `json:"type"     db:"type"` // cash, card, deposit
	Balance float64 `json:"balance"  db:"balance"`
}

// model/user.go
type User struct {
	BaseModel
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password_hash"` // никогда не сериализуем в JSON
	Name     string `json:"name" db:"name"`
}

// model/transaction.go
type Transaction struct {
	BaseModel
	UserID      int64     `json:"user_id" db:"user_id"`
	AccountID   int64     `json:"account_id" db:"account_id"`
	CategoryID  int64     `json:"category_id" db:"category_id"`
	Type        string    `json:"type" db:"type"` // "income" или "expense"
	Amount      float64   `json:"amount" db:"amount"`
	Date        time.Time `json:"date" db:"date"`
	Description string    `json:"description,omitempty" db:"description"`
}
