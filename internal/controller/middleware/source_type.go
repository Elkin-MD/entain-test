package middleware

import (
	"net/http"

	"entaintest/internal/common/enum"
)

// RequireSourceType rejects requests whose Source-Type header is missing or unsupported.
func RequireSourceType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !enum.SourceType(r.Header.Get("Source-Type")).Valid() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid or missing Source-Type header"}`))

			return
		}

		next.ServeHTTP(w, r)
	})
}
