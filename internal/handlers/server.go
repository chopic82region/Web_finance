package handlers

import (
	"context"
	"database/sql"
	"finance_tracker/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type StatusWriter struct {
	http.ResponseWriter
	status int
}

func (w *StatusWriter) writeHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

type middleware func(http.Handler) http.Handler

func chain(h http.Handler, m ...middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &StatusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		dur := time.Since(start)

		log.Printf("method=%s path=%s status=%d duration_ms=%d remote=%s ua=%q",
			r.Method,
			r.URL.Path,
			sw.status,
			dur.Milliseconds(),
			r.RemoteAddr,
			r.UserAgent(),
		)
	})
}

func NewRouter(db *sql.DB) http.Handler {
	h := NewHandler(
		service.NewUserRepo(db),
		service.NewAccountRepo(db),
		service.NewTransactionRepo(db),
	)

	mux := http.NewServeMux()

	// Healthcheck
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Users
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)

	// Accounts
	mux.HandleFunc("POST /accounts", h.CreateAccount)
	mux.HandleFunc("GET /accounts/{id}", h.GetAccountByID)
	mux.HandleFunc("PUT /accounts/{id}", h.UpdateAccount)
	mux.HandleFunc("DELETE /accounts/{id}", h.DeleteAccount)
	mux.HandleFunc("GET /users/{userId}/accounts", h.GetAccountsByUserID)

	// Transactions
	mux.HandleFunc("POST /transactions", h.CreateTransaction)
	mux.HandleFunc("GET /transactions/{id}", h.GetTransactionByID)
	mux.HandleFunc("PUT /transactions/{id}", h.UpdateTransaction)
	mux.HandleFunc("DELETE /transactions/{id}", h.DeleteTransaction)
	mux.HandleFunc("GET /users/{userId}/transactions", h.GetTransactionsByUserID)
	mux.HandleFunc("GET /accounts/{accountId}/transactions", h.GetTransactionsByAccountID)
	mux.HandleFunc("GET /users/{userId}/transactions/range", h.GetTransactionsByDateRange)

	return chain(mux, RequestLogger)
}

func StartServer() error {
	addr := os.Getenv("SERVER_PORT")
	db := database.NewDB()
	if err := db.Ping(); err != nil {
		return err
	}
	srv := &http.Server{
		Addr:    addr,
		Handler: NewRouter(db),
		// sane timeouts for REST API
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run server in background and wait for OS signal.
	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", addr)
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("shutdown signal: %s", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	case err := <-errCh:
		// http.ErrServerClosed is expected after Shutdown.
		if err == nil || err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
