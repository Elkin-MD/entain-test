package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/model/request"
	"entaintest/internal/model/response"
)

// Transaction processes POST /user/{userId}/transaction.
func (c *Controller) Transaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	var req request.TransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")

		return
	}

	req.UserID = userID
	req.SourceType = r.Header.Get("Source-Type")

	if _, err := c.svc.ProcessTransaction(r.Context(), req); err != nil {
		writeTransactionError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, response.TransactionResponse{Status: "ok"})
}

func writeTransactionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, businesserror.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid request")
	case errors.Is(err, businesserror.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, businesserror.ErrInsufficientFunds):
		writeError(w, http.StatusBadRequest, "insufficient funds")
	case errors.Is(err, businesserror.ErrDuplicateTransaction):
		writeError(w, http.StatusUnprocessableEntity, "transaction already processed")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
