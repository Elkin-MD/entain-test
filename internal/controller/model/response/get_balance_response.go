package response

import (
	"entaintest/internal/money"
	serviceresponse "entaintest/internal/service/model/response"
)

// GetBalanceResponse is the HTTP body returned for a balance lookup.
type GetBalanceResponse struct {
	UserID  uint64 `json:"userId"`
	Balance string `json:"balance"`
}

// NewGetBalanceResponse maps the service response to the HTTP response.
func NewGetBalanceResponse(response *serviceresponse.GetBalanceResponse) *GetBalanceResponse {
	return &GetBalanceResponse{
		UserID:  response.UserID,
		Balance: money.Format(response.Balance),
	}
}
