package controller

import (
	"net/http"
	"strconv"

	"entaintest/internal/controller/middleware"
	"entaintest/internal/service"
)

// Controller wires HTTP routes to the wallet service.
type Controller struct {
	svc *service.WalletService
}

// New creates a Controller over the given service.
func New(svc *service.WalletService) *Controller {
	return &Controller{svc: svc}
}

// Router builds the application HTTP handler with all routes and middleware.
func (c *Controller) Router() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /user/{userId}/transaction",
		middleware.RequireSourceType(http.HandlerFunc(c.Transaction)))

	mux.HandleFunc("GET /user/{userId}/balance", c.Balance)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return middleware.Recover(middleware.Logging(mux))
}

func parseUserID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(r.PathValue("userId"), 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, "userId must be a positive integer")

		return 0, false
	}

	return id, true
}
