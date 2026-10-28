package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   any    `json:"error"`
}

func JSON(w http.ResponseWriter, statusCode int, env Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(env)
}

func Success(w http.ResponseWriter, statusCode int, message string, data any) {
	JSON(w, statusCode, Envelope{
		Success: true,
		Message: message,
		Data:    data,
		Error:   nil,
	})
}

func Error(w http.ResponseWriter, statusCode int, message string, errDetails any) {
	JSON(w, statusCode, Envelope{
		Success: false,
		Message: message,
		Data:    nil,
		Error:   errDetails,
	})
}

func FromAppError(w http.ResponseWriter, err error) {
	Error(w, http.StatusBadRequest, "Request failed", err.Error())
}
