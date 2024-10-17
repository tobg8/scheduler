package helpers

import (
	"encoding/json"
	"net/http"

	models "github.com/tobg/scheduler/models/helpers"
)

// JSONResponse is a generic function to send JSON responses
func jsonResponse(w http.ResponseWriter, status int, r models.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(r)
}

// SendResponseMessage sends a response with a message
func SendResponseMessage(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, models.Response{Message: message, Status: status})
}

// SendResponse sends a response with data
func SendResponseData(w http.ResponseWriter, status int, data any) {
	jsonResponse(w, status, models.Response{Data: data, Status: status})
}
