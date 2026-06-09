package controller

import (
	"net/http"
	"strconv"
)

// Controller adapts HTTP requests to the wallet services.
type Controller struct {
	transactionService TransactionService
	balanceService     BalanceService
}

// New creates a Controller over the given services.
func New(transactionService TransactionService, balanceService BalanceService) *Controller {
	return &Controller{
		transactionService: transactionService,
		balanceService:     balanceService,
	}
}

func parseUserID(r *http.Request) (uint64, error) {
	id, err := strconv.ParseUint(r.PathValue("userId"), 10, 64)
	if err != nil || id == 0 {
		return 0, ErrInvalidUserID
	}

	return id, nil
}
