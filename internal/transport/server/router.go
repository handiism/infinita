package server

import (
	"net/http"
)

// NewHealthRouter creates a mux with endpoints that don't require auth (no API key).
func NewHealthRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	return mux
}

// NewAPIRouter creates a mux with all routes except /health (these require auth when apiKey is set).
func NewAPIRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /transactions", h.AddTransaction)
	mux.HandleFunc("GET /transactions", h.ListTransactions)

	mux.HandleFunc("GET /categories", h.ListCategories)
	mux.HandleFunc("POST /categories", h.CreateCategory)

	mux.HandleFunc("PUT /budgets", h.SetBudget)
	mux.HandleFunc("GET /budgets/status", h.GetBudgetStatus)

	mux.HandleFunc("GET /reports/daily", h.GetDailyReport)
	mux.HandleFunc("GET /reports/monthly", h.GetMonthlyReport)

	mux.HandleFunc("PUT /initial-balance", h.SetInitialBalance)
	mux.HandleFunc("DELETE /initial-balance", h.ResetInitialBalance)

	return mux
}

// NewRouter creates a mux with all routes including /health.
// Deprecated: Use NewHealthRouter and NewAPIRouter separately for proper auth handling.
func NewRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)

	mux.HandleFunc("POST /transactions", h.AddTransaction)
	mux.HandleFunc("GET /transactions", h.ListTransactions)

	mux.HandleFunc("GET /categories", h.ListCategories)
	mux.HandleFunc("POST /categories", h.CreateCategory)

	mux.HandleFunc("PUT /budgets", h.SetBudget)
	mux.HandleFunc("GET /budgets/status", h.GetBudgetStatus)

	mux.HandleFunc("GET /reports/daily", h.GetDailyReport)
	mux.HandleFunc("GET /reports/monthly", h.GetMonthlyReport)

	mux.HandleFunc("PUT /initial-balance", h.SetInitialBalance)
	mux.HandleFunc("DELETE /initial-balance", h.ResetInitialBalance)

	return mux
}
