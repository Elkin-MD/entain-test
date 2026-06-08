package request

import servicerequest "entaintest/internal/service/model/request"

// CreateTransactionRequest is the HTTP body for a transaction. UserID and
// SourceType come from the path and headers, not the JSON body.
type CreateTransactionRequest struct {
	UserID        uint64 `json:"-"`
	SourceType    string `json:"-"`
	State         string `json:"state"`
	Amount        string `json:"amount"`
	TransactionID string `json:"transactionId"`
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
