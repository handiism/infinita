package server

import (
	"net/http"
)

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

	mux.HandleFunc("GET /settings", h.GetSettings)

	mux.HandleFunc("PUT /settings/initial-balance", h.SetInitialBalance)
	mux.HandleFunc("DELETE /settings/initial-balance", h.ResetInitialBalance)

	mux.HandleFunc("PUT /settings/report-timezone", h.SetReportTimezone)

	return mux
}
