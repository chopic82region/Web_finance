package apperrors

import "errors"

var ErrCreateUser = errors.New("error creating user")
var ErrInvalidEmail = errors.New("invalid email")

var ErrUserNotFound = errors.New("user not found")
var ErrDuplicateEmail = errors.New("email already exists")

var ErrAccountNotFound = errors.New("account not found")
var ErrTransactionNotFound = errors.New("transaction not found")

