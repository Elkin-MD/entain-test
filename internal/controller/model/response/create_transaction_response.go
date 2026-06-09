package response

import (
	"entaintest/internal/money"
	serviceresponse "entaintest/internal/service/model/response"
)

// CreateTransactionResponse is the HTTP body returned for a transaction.
type CreateTransactionResponse struct {
	TransactionID string `json:"transactionId"`
	UserID        uint64 `json:"userId"`
	State         string `json:"state"`
	SourceType    string `json:"sourceType"`
	Amount        string `json:"amount"`
	Balance       string `json:"balance"`
}

// NewCreateTransactionResponse maps the service response to the HTTP response.
func NewCreateTransactionResponse(response *serviceresponse.ProcessTransactionResponse) *CreateTransactionResponse {
	return &CreateTransactionResponse{
		TransactionID: response.TransactionID,
		UserID:        response.UserID,
		State:         string(response.State),
		SourceType:    string(response.SourceType),
		Amount:        money.Format(response.Amount),
		Balance:       money.Format(response.Balance),
	}
}
