package request

import (
	"entaintest/internal/common/businesserror"
	"entaintest/internal/common/enum"
	"entaintest/internal/money"
	servicerequest "entaintest/internal/service/model/request"
)

// CreateTransactionRequest is the HTTP body for a transaction. UserID and
// SourceType come from the path and headers, not the JSON body.
type CreateTransactionRequest struct {
	UserID        uint64 `json:"-"`
	SourceType    string `json:"-"`
	State         string `json:"state"`
	Amount        string `json:"amount"`
	TransactionID string `json:"transactionId"`
}

// NewCreateTransactionRequest validates the inputs and builds the request.
func NewCreateTransactionRequest(userID uint64, sourceType, state, amount, transactionID string) (CreateTransactionRequest, error) {
	request := CreateTransactionRequest{
		UserID:        userID,
		SourceType:    sourceType,
		State:         state,
		Amount:        amount,
		TransactionID: transactionID,
	}

	if err := request.validate(); err != nil {
		return CreateTransactionRequest{}, err
	}

	return request, nil
}

func (r CreateTransactionRequest) validate() error {
	if !enum.TransactionState(r.State).Valid() {
		return businesserror.ErrInvalidState
	}

	if !enum.SourceType(r.SourceType).Valid() {
		return businesserror.ErrInvalidSourceType
	}

	if r.TransactionID == "" {
		return businesserror.ErrMissingTransactionID
	}

	if _, err := money.Parse(r.Amount); err != nil {
		return businesserror.ErrInvalidAmount
	}

	return nil
}

// ToServiceRequest maps the HTTP request to the service request.
func (r CreateTransactionRequest) ToServiceRequest() servicerequest.ProcessTransactionRequest {
	return servicerequest.ProcessTransactionRequest{
		UserID:        r.UserID,
		SourceType:    r.SourceType,
		State:         r.State,
		Amount:        r.Amount,
		TransactionID: r.TransactionID,
	}
}
