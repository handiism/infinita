package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/handiism/infinita/internal/application/port/input"
	domainerror "github.com/handiism/infinita/internal/domain/error"
)

type Handler struct {
	Transactions input.TransactionUseCase
	Categories   input.CategoryUseCase
	Budgets      input.BudgetUseCase
	Reports      input.ReportUseCase
	Settings     input.SettingsUseCase
}

const mvpCurrencyCode = "IDR"

func NewHandler(
	transactions input.TransactionUseCase,
	categories input.CategoryUseCase,
	budgets input.BudgetUseCase,
	reports input.ReportUseCase,
	settings input.SettingsUseCase,
) *Handler {
	return &Handler{
		Transactions: transactions,
		Categories:   categories,
		Budgets:      budgets,
		Reports:      reports,
		Settings:     settings,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	WriteSuccess(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (h *Handler) AddTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	var errs []domainerror.DomainError
	if req.Type != "income" && req.Type != "expense" {
		errs = append(errs, domainerror.ErrInvalidTransactionType.WithField("type"))
	}
	if req.AmountMinor <= 0 {
		errs = append(errs, domainerror.ErrInvalidAmount.WithField("amountMinor"))
	}
	if req.Category == "" {
		errs = append(errs, domainerror.ErrInvalidCategory.WithField("category"))
	}
	if req.Date == "" {
		errs = append(errs, domainerror.ErrInvalidDate.WithField("date"))
	}
	if req.CurrencyCode != mvpCurrencyCode {
		errs = append(errs, domainerror.ErrInvalidCurrency.WithField("currencyCode"))
	}
	if len(errs) > 0 {
		WriteFail(w, http.StatusBadRequest, errs)
		return
	}

	txn, err := h.Transactions.AddTransaction(r.Context(), req.Type, req.AmountMinor, req.Category, req.Date, req.Description)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusCreated, TransactionResponse{
		ID:                   txn.ID,
		Type:                 txn.Type,
		AmountMinor:          txn.AmountMinor,
		CurrencyCode:         txn.CurrencyCode,
		CategoryID:           txn.CategoryID,
		CategoryNameSnapshot: txn.CategoryNameSnapshot,
		Date:                 txn.Date,
		Description:          txn.Description,
		CreatedAt:            txn.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 1 || parsed > 500 {
			WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
				domainerror.ErrInvalidFlag.WithField("limit").WithHint("must be an integer between 1 and 500"),
			})
			return
		}
		limit = parsed
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		parsed, err := strconv.Atoi(o)
		if err != nil || parsed < 0 {
			WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
				domainerror.ErrInvalidFlag.WithField("offset").WithHint("must be a non-negative integer"),
			})
			return
		}
		offset = parsed
	}

	result, err := h.Transactions.ListTransactions(r.Context(), categoryPtr, limit, offset)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	resp := make([]TransactionResponse, len(result.Transactions))
	for i, t := range result.Transactions {
		resp[i] = TransactionResponse{
			ID:                   t.ID,
			Type:                 t.Type,
			AmountMinor:          t.AmountMinor,
			CurrencyCode:         t.CurrencyCode,
			CategoryID:           t.CategoryID,
			CategoryNameSnapshot: t.CategoryNameSnapshot,
			Date:                 t.Date,
			Description:          t.Description,
			CreatedAt:            t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	WriteSuccessWithMeta(w, http.StatusOK, resp, PaginationMeta{
		Total:  result.Total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Categories.List(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	resp := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		resp[i] = CategoryResponse{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
		}
	}

	WriteSuccess(w, http.StatusOK, resp)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	if req.Name == "" {
		WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
			domainerror.ErrInvalidCategory.WithField("name"),
		})
		return
	}

	category, err := h.Categories.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusCreated, CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
	})
}

func (h *Handler) SetBudget(w http.ResponseWriter, r *http.Request) {
	var req BudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	var errs []domainerror.DomainError
	if req.Category == "" {
		errs = append(errs, domainerror.ErrInvalidCategory.WithField("category"))
	}
	if req.Month == "" {
		errs = append(errs, domainerror.ErrInvalidMonth.WithField("month"))
	}
	if req.MonthlyLimitMinor <= 0 {
		errs = append(errs, domainerror.ErrInvalidAmount.WithField("monthlyLimitMinor"))
	}
	if len(errs) > 0 {
		WriteFail(w, http.StatusBadRequest, errs)
		return
	}

	err := h.Budgets.SetBudget(r.Context(), req.Category, req.Month, req.MonthlyLimitMinor)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, nil)
}

func (h *Handler) GetBudgetStatus(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
			domainerror.ErrInvalidMonth.WithField("month"),
		})
		return
	}

	statuses, err := h.Budgets.Status(r.Context(), month)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	resp := make([]BudgetStatusResponse, len(statuses))
	for i, s := range statuses {
		resp[i] = BudgetStatusResponse{
			CategoryName:          s.CategoryName,
			CurrencyCode:          s.CurrencyCode,
			MonthlyLimitMinor:     s.MonthlyLimitMinor,
			SpentMonthToDateMinor: s.SpentMonthToDateMinor,
			RemainingMinor:        s.RemainingMinor,
			IsOverLimit:           s.IsOverLimit,
		}
	}

	WriteSuccess(w, http.StatusOK, resp)
}

func (h *Handler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
			domainerror.ErrInvalidDate.WithField("date"),
		})
		return
	}

	summary, err := h.Reports.Daily(r.Context(), date)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, DailyReportResponse{
		Period:            "daily",
		Date:              summary.Period,
		CurrencyCode:      summary.CurrencyCode,
		IncomeTotalMinor:  summary.IncomeTotalMinor,
		ExpenseTotalMinor: summary.ExpenseTotalMinor,
		NetBalanceMinor:   summary.NetBalanceMinor,
	})
}

func (h *Handler) GetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
			domainerror.ErrInvalidMonth.WithField("month"),
		})
		return
	}

	summary, err := h.Reports.Monthly(r.Context(), month)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	topCategories := make([]TopCategoryResponse, len(summary.TopCategories))
	for i, tc := range summary.TopCategories {
		topCategories[i] = TopCategoryResponse{
			Category:    tc.Category,
			AmountMinor: tc.AmountMinor,
		}
	}

	WriteSuccess(w, http.StatusOK, MonthlyReportResponse{
		Period:              "monthly",
		Month:               summary.Period,
		CurrencyCode:        summary.CurrencyCode,
		IncomeTotalMinor:    summary.IncomeTotalMinor,
		ExpenseTotalMinor:   summary.ExpenseTotalMinor,
		NetBalanceMinor:     summary.NetBalanceMinor,
		ClosingBalanceMinor: summary.ClosingBalanceMinor,
		TopCategories:       topCategories,
	})
}

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Settings.Show(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, SettingsResponse{
		StorageMode:    settings.StorageMode,
		AnalyticsOptIn: settings.AnalyticsOptIn,
		ReportTimezone: settings.ReportTimezone,
	})
}

func (h *Handler) SetInitialBalance(w http.ResponseWriter, r *http.Request) {
	var req SetInitialBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	if req.InitialBalanceMinor < 0 {
		WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
			domainerror.ErrInvalidAmount.WithField("initialBalanceMinor").WithHint("provide a non-negative numeric value"),
		})
		return
	}
	if req.CurrencyCode != mvpCurrencyCode {
		WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
			domainerror.ErrInvalidCurrency.WithField("currencyCode"),
		})
		return
	}

	initialBalance, err := h.Settings.SetInitialBalance(r.Context(), req.InitialBalanceMinor)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, InitialBalanceResponse{
		InitialBalanceMinor: initialBalance.InitialBalanceMinor,
		CurrencyCode:        initialBalance.CurrencyCode,
		InitializedAt:       initialBalance.InitializedAt,
	})
}

func (h *Handler) ResetInitialBalance(w http.ResponseWriter, r *http.Request) {
	err := h.Settings.ResetInitialBalance(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, nil)
}

func (h *Handler) SetAnalyticsOptIn(w http.ResponseWriter, r *http.Request) {
	var req SetAnalyticsOptInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	err := h.Settings.SetAnalyticsOptIn(r.Context(), req.AnalyticsOptIn)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, nil)
}

func (h *Handler) SetReportTimezone(w http.ResponseWriter, r *http.Request) {
	var req SetReportTimezoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	if req.ReportTimezone == "" {
		WriteFail(w, http.StatusBadRequest, []domainerror.DomainError{
			domainerror.ErrInvalidTimezone.WithField("reportTimezone"),
		})
		return
	}

	err := h.Settings.SetReportTimezone(r.Context(), req.ReportTimezone)
	if err != nil {
		if WriteFailFromError(w, http.StatusBadRequest, err) {
			return
		}
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, nil)
}
