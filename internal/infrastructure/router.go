package infrastructure

import (
	"net/http"

	"entaintest/internal/controller"
	"entaintest/internal/infrastructure/middleware"
)

// NewRouter builds the application HTTP handler with all routes and middleware.
func NewRouter(c *controller.Controller) http.Handler {
	mux := http.NewServeMux()

	// POST is used to create a transaction; PUT would be more conventional for an upsert.
	mux.Handle("POST /user/{userId}/transaction", middleware.RequireSourceType(handle(c.CreateTransaction)))
	mux.Handle("GET /user/{userId}/balance", handle(c.GetBalance))
	mux.HandleFunc("GET /health", health)

	return middleware.Recover(middleware.Logging(mux))
}

// handle adapts a controller method that returns a typed response into an
// http.Handler, writing 200 with the JSON body or mapping the error to a status.
func handle[T any](next func(http.ResponseWriter, *http.Request) (*T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := next(w, r)
		if err != nil {
			writeError(w, err)

			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
