package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"finance_tracker/internal/apperrors"
	"finance_tracker/internal/model"
	"finance_tracker/internal/service"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	users        *service.UserRepository
	accounts     *service.AccountRepository
	transactions *service.TransactionRepository
}

func NewHandler(users *service.UserRepository, accounts *service.AccountRepository, transactions *service.TransactionRepository) *Handler {
	return &Handler{
		users:        users,
		accounts:     accounts,
		transactions: transactions,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	if err == nil {
		writeJSON(w, status, errorResponse{Error: http.StatusText(status)})
		return
	}
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func parseIDParam(r *http.Request, name string) (int64, error) {
	raw := r.PathValue(name)
	if raw == "" {
		return 0, errors.New("missing id")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func mapRepoError(err error) (status int, body errorResponse) {
	switch {
	case err == nil:
		return http.StatusOK, errorResponse{}
	case errors.Is(err, apperrors.ErrInvalidEmail):
		return http.StatusBadRequest, errorResponse{Error: err.Error()}
	case errors.Is(err, apperrors.ErrDuplicateEmail):
		return http.StatusConflict, errorResponse{Error: err.Error()}
	case errors.Is(err, apperrors.ErrUserNotFound),
		errors.Is(err, apperrors.ErrAccountNotFound),
		errors.Is(err, apperrors.ErrTransactionNotFound):
		return http.StatusNotFound, errorResponse{Error: err.Error()}
	default:
		return http.StatusInternalServerError, errorResponse{Error: "internal server error"}
	}
}

// ---- Users ----

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	u := &model.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}
	if err := h.users.Create(r.Context(), u); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	writeJSON(w, http.StatusCreated, userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
}

type userResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type updateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req updateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	u := &model.User{
		BaseModel: model.BaseModel{ID: id},
		Email:     req.Email,
		Password:  req.Password,
		Name:      req.Name,
	}
	if err := h.users.Update(r.Context(), u); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	writeJSON(w, http.StatusOK, userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	u := &model.User{BaseModel: model.BaseModel{ID: id}}
	if err := h.users.Delete(r.Context(), u); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- Accounts ----

type createAccountRequest struct {
	UserID  int64   `json:"user_id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Balance float64 `json:"balance"`
}

type accountResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toAccountResponse(a *model.Account) accountResponse {
	return accountResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Name:      a.Name,
		Type:      a.Type,
		Balance:   a.Balance,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	a := &model.Account{
		UserID:  req.UserID,
		Name:    req.Name,
		Type:    req.Type,
		Balance: req.Balance,
	}
	if err := h.accounts.Create(r.Context(), a); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	writeJSON(w, http.StatusCreated, toAccountResponse(a))
}

func (h *Handler) GetAccountByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	a, err := h.accounts.GetByID(r.Context(), id)
	if err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	writeJSON(w, http.StatusOK, toAccountResponse(a))
}

func (h *Handler) GetAccountsByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "userId")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	accounts, err := h.accounts.GetByUserID(r.Context(), userID)
	if err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	out := make([]accountResponse, 0, len(accounts))
	for i := range accounts {
		a := accounts[i]
		out = append(out, toAccountResponse(&a))
	}
	writeJSON(w, http.StatusOK, out)
}

type updateAccountRequest struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Balance float64 `json:"balance"`
}

func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req updateAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	a := &model.Account{
		BaseModel: model.BaseModel{ID: id},
		Name:      req.Name,
		Type:      req.Type,
		Balance:   req.Balance,
	}
	if err := h.accounts.Update(r.Context(), a); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	writeJSON(w, http.StatusOK, toAccountResponse(a))
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.accounts.Delete(r.Context(), id); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- Transactions ----

type createTransactionRequest struct {
	UserID      int64     `json:"user_id"`
	AccountID   int64     `json:"account_id"`
	CategoryID  int64     `json:"category_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type transactionResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	AccountID   int64     `json:"account_id"`
	CategoryID  int64     `json:"category_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toTransactionResponse(t *model.Transaction) transactionResponse {
	return transactionResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		AccountID:   t.AccountID,
		CategoryID:  t.CategoryID,
		Type:        t.Type,
		Amount:      t.Amount,
		Date:        t.Date,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req createTransactionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	t := &model.Transaction{
		UserID:      req.UserID,
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Type:        req.Type,
		Amount:      req.Amount,
		Date:        req.Date,
		Description: req.Description,
	}
	if err := h.transactions.Create(r.Context(), t); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	writeJSON(w, http.StatusCreated, toTransactionResponse(t))
}

func (h *Handler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	t, err := h.transactions.GetByID(r.Context(), id)
	if err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	writeJSON(w, http.StatusOK, toTransactionResponse(t))
}

func (h *Handler) GetTransactionsByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "userId")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	txns, err := h.transactions.GetByUserID(r.Context(), userID)
	if err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	out := make([]transactionResponse, 0, len(txns))
	for i := range txns {
		t := txns[i]
		out = append(out, toTransactionResponse(&t))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) GetTransactionsByAccountID(w http.ResponseWriter, r *http.Request) {
	accountID, err := parseIDParam(r, "accountId")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	txns, err := h.transactions.GetByAccountID(r.Context(), accountID)
	if err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	out := make([]transactionResponse, 0, len(txns))
	for i := range txns {
		t := txns[i]
		out = append(out, toTransactionResponse(&t))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) GetTransactionsByDateRange(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "userId")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	fromRaw := r.URL.Query().Get("from")
	toRaw := r.URL.Query().Get("to")
	if fromRaw == "" || toRaw == "" {
		writeError(w, http.StatusBadRequest, errors.New("from and to are required"))
		return
	}
	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid from (use RFC3339)"))
		return
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid to (use RFC3339)"))
		return
	}

	txns, err := h.transactions.GetByDateRange(r.Context(), userID, from, to)
	if err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	out := make([]transactionResponse, 0, len(txns))
	for i := range txns {
		t := txns[i]
		out = append(out, toTransactionResponse(&t))
	}
	writeJSON(w, http.StatusOK, out)
}

type updateTransactionRequest struct {
	AccountID   int64     `json:"account_id"`
	CategoryID  int64     `json:"category_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

func (h *Handler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var req updateTransactionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	t := &model.Transaction{
		BaseModel:   model.BaseModel{ID: id},
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Type:        req.Type,
		Amount:      req.Amount,
		Date:        req.Date,
		Description: req.Description,
	}
	if err := h.transactions.Update(r.Context(), t); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	writeJSON(w, http.StatusOK, toTransactionResponse(t))
}

func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.transactions.Delete(r.Context(), id); err != nil {
		status, body := mapRepoError(err)
		writeJSON(w, status, body)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// withTimeout is optional, but keeps handlers responsive under DB issues.
func withTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}
