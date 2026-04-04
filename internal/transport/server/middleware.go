package server

import (
	"encoding/json"
	"net/http"
)

// APIKeyMiddleware validates the X-API-Key header when apiKey is non-empty.
// When apiKey is empty, the middleware passes through without authentication.
// The /health endpoint is handled separately and does not go through this middleware.
func APIKeyMiddleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("X-API-Key") != apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "fail",
				"data": []map[string]string{{
					"code":    "UNAUTHORIZED",
					"message": "invalid or missing API key",
				}},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
