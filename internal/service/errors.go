package service

import "finance_tracker/internal/apperrors"

// Backward-compat re-exports (чтобы остальной код мог импортировать service.*).
var (
	ErrOfCreateUser       = apperrors.ErrCreateUser
	ErrInvalidEmail       = apperrors.ErrInvalidEmail
	ErrUserNotFound       = apperrors.ErrUserNotFound
	ErrDuplicateEmail     = apperrors.ErrDuplicateEmail
	ErrAccountNotFound    = apperrors.ErrAccountNotFound
	ErrTransactionNotFound = apperrors.ErrTransactionNotFound
)

