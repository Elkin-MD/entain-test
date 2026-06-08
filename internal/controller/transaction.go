package controller

import (
	"context"
	"encoding/json"
	"net/http"

	controllerrequest "entaintest/internal/controller/model/request"
	controllerresponse "entaintest/internal/controller/model/response"
	servicerequest "entaintest/internal/service/model/request"
	serviceresponse "entaintest/internal/service/model/response"
)

// TransactionService is the transaction behaviour the controller depends on.
type TransactionService interface {
	ProcessTransaction(ctx context.Context, request servicerequest.ProcessTransactionRequest) (*serviceresponse.ProcessTransactionResponse, error)
}

// CreateTransaction handles POST /user/{userId}/transaction.
func (c *Controller) CreateTransaction(_ http.ResponseWriter, r *http.Request) (*controllerresponse.CreateTransactionResponse, error) {
	userID, err := parseUserID(r)
	if err != nil {
		return nil, err
	}

	var request controllerrequest.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, ErrInvalidBody
	}

	request.UserID = userID
	request.SourceType = r.Header.Get("Source-Type")

	response, err := c.transactionService.ProcessTransaction(r.Context(), request.ToServiceRequest())
	if err != nil {
		return nil, err
	}

	return controllerresponse.NewCreateTransactionResponse(response), nil
}
