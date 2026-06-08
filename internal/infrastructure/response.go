package infrastructure

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/controller"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if body == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, statusForError(err), errorResponse{Error: err.Error()})
}

func statusForError(err error) int {
	var businessErr businesserror.Error
	var requestErr controller.RequestError

	switch {
	case errors.As(err, &requestErr):
		return http.StatusBadRequest
	case errors.As(err, &businessErr):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
