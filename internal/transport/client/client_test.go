package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientListTransactionsDecodesJSendDataAndMeta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/transactions", r.URL.Path)
		require.Equal(t, "10", r.URL.Query().Get("limit"))
		require.Equal(t, "5", r.URL.Query().Get("offset"))

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{
					"id":                   "txn-1",
					"type":                 "expense",
					"amountMinor":          int64(12345),
					"currencyCode":         "IDR",
					"categoryId":           int64(1),
					"categoryNameSnapshot": "Food",
					"date":                 "2024-01-15",
					"description":          "Lunch",
					"createdAt":            "2024-01-15T10:00:00Z",
				},
			},
			"meta": map[string]interface{}{
				"total":  42,
				"limit":  10,
				"offset": 5,
			},
		}))
	}))
	defer server.Close()

	client := New(server.URL)
	result, err := client.ListTransactions(context.Background(), nil, 10, 5)

	require.NoError(t, err)
	require.Len(t, result.Transactions, 1)
	require.Equal(t, "txn-1", result.Transactions[0].ID)
	require.Equal(t, 42, result.Total)
}

func TestClientListDecodesJSendArrayData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/categories", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{
					"id":          int64(1),
					"name":        "Food",
					"description": "Meals",
				},
			},
		}))
	}))
	defer server.Close()

	client := New(server.URL)
	result, err := client.List(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Food", result[0].Name)
}

func TestClientStatusDecodesJSendArrayData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/budgets/status", r.URL.Path)
		require.Equal(t, "2024-01", r.URL.Query().Get("month"))

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{
					"categoryName":          "Food",
					"currencyCode":          "IDR",
					"monthlyLimitMinor":     int64(50000),
					"spentMonthToDateMinor": int64(12000),
					"remainingMinor":        int64(38000),
					"isOverLimit":           false,
				},
			},
		}))
	}))
	defer server.Close()

	client := New(server.URL)
	result, err := client.Status(context.Background(), "2024-01")

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Food", result[0].CategoryName)
	require.Equal(t, int64(38000), result[0].RemainingMinor)
}
