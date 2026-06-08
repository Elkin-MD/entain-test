package controller

import (
	"context"
	"net/http"

	controllerresponse "entaintest/internal/controller/model/response"
	serviceresponse "entaintest/internal/service/model/response"
)

// BalanceService is the balance behaviour the controller depends on.
type BalanceService interface {
	GetBalance(ctx context.Context, userID uint64) (*serviceresponse.GetBalanceResponse, error)
}

// GetBalance handles GET /user/{userId}/balance.
func (c *Controller) GetBalance(_ http.ResponseWriter, r *http.Request) (*controllerresponse.GetBalanceResponse, error) {
	userID, err := parseUserID(r)
	if err != nil {
		return nil, err
	}

	response, err := c.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		return nil, err
	}

	return controllerresponse.NewGetBalanceResponse(response), nil
}
