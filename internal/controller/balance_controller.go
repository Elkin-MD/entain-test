package controller

import (
	"errors"
	"net/http"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/model/response"
	"entaintest/internal/money"
)

// Balance processes GET /user/{userId}/balance.
func (c *Controller) Balance(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	balance, err := c.svc.GetBalance(r.Context(), userID)
	if err != nil {
		if errors.Is(err, businesserror.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")

			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeJSON(w, http.StatusOK, response.BalanceResponse{
		UserID:  userID,
		Balance: money.Format(balance),
	})
}
