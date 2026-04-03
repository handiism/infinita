package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/handiism/infinita/internal/application/port/input"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) performRequest(ctx context.Context, method, path string, body interface{}) (Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return Response{}, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, &ClientError{Code: "CONNECTION_ERROR", Message: err.Error()}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("client performRequest: failed to close response body: %v", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read response body: %w", err)
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return Response{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return Response{}, c.handleError(response, resp.StatusCode)
	}

	if response.Status != "success" {
		return Response{}, fmt.Errorf("unexpected response status %q for HTTP %d", response.Status, resp.StatusCode)
	}

	return response, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	response, err := c.performRequest(ctx, method, path, body)
	if err != nil {
		return err
	}

	if result != nil && response.Data != nil {
		dataBytes, err := json.Marshal(response.Data)
		if err != nil {
			return fmt.Errorf("remarshal data: %w", err)
		}
		if err := json.Unmarshal(dataBytes, result); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	}

	return nil
}

func (c *Client) doRequestWithMeta(ctx context.Context, method, path string, body interface{}, result interface{}, meta interface{}) error {
	response, err := c.performRequest(ctx, method, path, body)
	if err != nil {
		return err
	}

	if result != nil && response.Data != nil {
		dataBytes, err := json.Marshal(response.Data)
		if err != nil {
			return fmt.Errorf("remarshal data: %w", err)
		}
		if err := json.Unmarshal(dataBytes, result); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	}

	if meta != nil && response.Meta != nil {
		metaBytes, err := json.Marshal(response.Meta)
		if err != nil {
			return fmt.Errorf("remarshal meta: %w", err)
		}
		if err := json.Unmarshal(metaBytes, meta); err != nil {
			return fmt.Errorf("unmarshal meta: %w", err)
		}
	}

	return nil
}

func (c *Client) handleError(response Response, statusCode int) error {
	if response.Status == "fail" && response.Data != nil {
		errorObjects, ok := response.Data.([]interface{})
		if !ok {
			return &ClientError{Code: "UNKNOWN_ERROR", Message: "unexpected error format"}
		}

		var errors []domainerror.DomainError
		for _, e := range errorObjects {
			errorMap, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			de := domainerror.DomainError{
				Code:    getString(errorMap, "code"),
				Message: getString(errorMap, "message"),
				Field:   getString(errorMap, "field"),
				Hint:    getString(errorMap, "hint"),
			}
			errors = append(errors, de)
		}

		if len(errors) == 0 {
			return &ClientError{Code: "UNKNOWN_ERROR", Message: "server returned validation failure with no parseable errors", StatusCode: statusCode}
		}
		if len(errors) == 1 {
			return &ClientError{DomainError: errors[0], StatusCode: statusCode}
		}
		return &ClientError{MultipleErrors: errors, StatusCode: statusCode}
	}

	if response.Status == "error" {
		return &ClientError{
			Code:       response.Code,
			Message:    response.Message,
			StatusCode: statusCode,
		}
	}

	return &ClientError{Code: "UNKNOWN_ERROR", Message: "unexpected error", StatusCode: statusCode}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func (c *Client) get(ctx context.Context, path string, query url.Values, result interface{}) error {
	fullURL := path
	if len(query) > 0 {
		fullURL = path + "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, fullURL, nil, result)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

func (c *Client) put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPut, path, body, result)
}

func (c *Client) delete(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, result)
}

func (c *Client) HealthCheck(ctx context.Context) error {
	return c.get(ctx, "/health", nil, nil)
}

func (c *Client) AddTransaction(ctx context.Context, entryType string, amountMinor int64, category string, date string, description string) (entity.Transaction, error) {
	req := TransactionRequest{
		Type:         entryType,
		AmountMinor:  amountMinor,
		CurrencyCode: "IDR",
		Category:     category,
		Date:         date,
		Description:  description,
	}
	var resp TransactionResponse
	if err := c.post(ctx, "/transactions", req, &resp); err != nil {
		return entity.Transaction{}, err
	}
	return entity.Transaction{
		ID:                   resp.ID,
		Type:                 resp.Type,
		AmountMinor:          resp.AmountMinor,
		CurrencyCode:         resp.CurrencyCode,
		CategoryID:           resp.CategoryID,
		CategoryNameSnapshot: resp.CategoryNameSnapshot,
		Date:                 resp.Date,
		Description:          resp.Description,
		CreatedAt:            resp.CreatedAt,
	}, nil
}

func (c *Client) ListTransactions(ctx context.Context, category *string, limit, offset int) (input.TransactionListResult, error) {
	query := url.Values{}
	if category != nil {
		query.Set("category", *category)
	}
	query.Set("limit", fmt.Sprintf("%d", limit))
	query.Set("offset", fmt.Sprintf("%d", offset))

	var resp TransactionListResponse
	var meta PaginationMeta
	if err := c.doRequestWithMeta(ctx, http.MethodGet, "/transactions?"+query.Encode(), nil, &resp, &meta); err != nil {
		return input.TransactionListResult{}, err
	}

	transactions := make([]entity.Transaction, len(resp))
	for i, t := range resp {
		transactions[i] = entity.Transaction{
			ID:                   t.ID,
			Type:                 t.Type,
			AmountMinor:          t.AmountMinor,
			CurrencyCode:         t.CurrencyCode,
			CategoryID:           t.CategoryID,
			CategoryNameSnapshot: t.CategoryNameSnapshot,
			Date:                 t.Date,
			Description:          t.Description,
			CreatedAt:            t.CreatedAt,
		}
	}
	return input.TransactionListResult{
		Transactions: transactions,
		Total:        meta.Total,
	}, nil
}

func (c *Client) List(ctx context.Context) ([]entity.Category, error) {
	var resp CategoryListResponse
	if err := c.get(ctx, "/categories", nil, &resp); err != nil {
		return nil, err
	}

	categories := make([]entity.Category, len(resp))
	for i, cat := range resp {
		categories[i] = entity.Category{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
		}
	}
	return categories, nil
}

func (c *Client) Create(ctx context.Context, name, description string) (entity.Category, error) {
	req := CategoryRequest{
		Name:        name,
		Description: description,
	}
	var resp CategoryResponse
	if err := c.post(ctx, "/categories", req, &resp); err != nil {
		return entity.Category{}, err
	}
	return entity.Category{
		ID:          resp.ID,
		Name:        resp.Name,
		Description: resp.Description,
	}, nil
}

func (c *Client) SetBudget(ctx context.Context, category, month string, limit int64) error {
	req := BudgetRequest{
		Category:          category,
		Month:             month,
		MonthlyLimitMinor: limit,
	}
	return c.put(ctx, "/budgets", req, nil)
}

func (c *Client) Status(ctx context.Context, month string) ([]entity.BudgetStatus, error) {
	query := url.Values{}
	query.Set("month", month)

	var resp BudgetStatusListResponse
	if err := c.get(ctx, "/budgets/status", query, &resp); err != nil {
		return nil, err
	}

	statuses := make([]entity.BudgetStatus, len(resp))
	for i, s := range resp {
		statuses[i] = entity.BudgetStatus{
			CategoryName:          s.CategoryName,
			CurrencyCode:          s.CurrencyCode,
			Month:                 month,
			MonthlyLimitMinor:     s.MonthlyLimitMinor,
			SpentMonthToDateMinor: s.SpentMonthToDateMinor,
			RemainingMinor:        s.RemainingMinor,
			IsOverLimit:           s.IsOverLimit,
		}
	}
	return statuses, nil
}

func (c *Client) Daily(ctx context.Context, date string) (entity.DailySummary, error) {
	query := url.Values{}
	query.Set("date", date)

	var resp DailyReportResponse
	if err := c.get(ctx, "/reports/daily", query, &resp); err != nil {
		return entity.DailySummary{}, err
	}
	return entity.DailySummary{
		Period:            resp.Date,
		CurrencyCode:      resp.CurrencyCode,
		IncomeTotalMinor:  resp.IncomeTotalMinor,
		ExpenseTotalMinor: resp.ExpenseTotalMinor,
		NetBalanceMinor:   resp.NetBalanceMinor,
	}, nil
}

func (c *Client) Monthly(ctx context.Context, month string) (entity.MonthlySummary, error) {
	query := url.Values{}
	query.Set("month", month)

	var resp MonthlyReportResponse
	if err := c.get(ctx, "/reports/monthly", query, &resp); err != nil {
		return entity.MonthlySummary{}, err
	}

	topCategories := make([]entity.TopSpendingCategory, len(resp.TopCategories))
	for i, tc := range resp.TopCategories {
		topCategories[i] = entity.TopSpendingCategory{
			Category:    tc.Category,
			AmountMinor: tc.AmountMinor,
		}
	}

	return entity.MonthlySummary{
		Period:              resp.Month,
		CurrencyCode:        resp.CurrencyCode,
		IncomeTotalMinor:    resp.IncomeTotalMinor,
		ExpenseTotalMinor:   resp.ExpenseTotalMinor,
		NetBalanceMinor:     resp.NetBalanceMinor,
		ClosingBalanceMinor: resp.ClosingBalanceMinor,
		TopCategories:       topCategories,
	}, nil
}

func (c *Client) Show(ctx context.Context) (entity.Settings, error) {
	var resp SettingsResponse
	if err := c.get(ctx, "/settings", nil, &resp); err != nil {
		return entity.Settings{}, err
	}
	return entity.Settings{
		StorageMode:    resp.StorageMode,
		AnalyticsOptIn: resp.AnalyticsOptIn,
		ReportTimezone: resp.ReportTimezone,
	}, nil
}

func (c *Client) SetInitialBalance(ctx context.Context, amount int64) (entity.InitialBalance, error) {
	req := SetInitialBalanceRequest{
		InitialBalanceMinor: amount,
		CurrencyCode:        "IDR",
	}
	var resp InitialBalanceResponse
	if err := c.put(ctx, "/settings/initial-balance", req, &resp); err != nil {
		return entity.InitialBalance{}, err
	}
	return entity.InitialBalance{
		InitialBalanceMinor: resp.InitialBalanceMinor,
		CurrencyCode:        resp.CurrencyCode,
		InitializedAt:       resp.InitializedAt,
	}, nil
}

func (c *Client) ResetInitialBalance(ctx context.Context) error {
	return c.delete(ctx, "/settings/initial-balance", nil)
}

func (c *Client) SetAnalyticsOptIn(ctx context.Context, optIn bool) error {
	req := SetAnalyticsOptInRequest{
		AnalyticsOptIn: optIn,
	}
	return c.put(ctx, "/settings/analytics", req, nil)
}

func (c *Client) SetReportTimezone(ctx context.Context, timezone string) error {
	req := SetReportTimezoneRequest{
		ReportTimezone: timezone,
	}
	return c.put(ctx, "/settings/report-timezone", req, nil)
}

var (
	_ input.TransactionUseCase = (*Client)(nil)
	_ input.CategoryUseCase    = (*Client)(nil)
	_ input.BudgetUseCase      = (*Client)(nil)
	_ input.ReportUseCase      = (*Client)(nil)
	_ input.SettingsUseCase    = (*Client)(nil)
)
