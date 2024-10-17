package models

// Response represents the structure of a successful response
type Response struct {
	Message string `json:"message,omitempty"`
	Status  int    `json:"status"`
	Data    any    `json:"data,omitempty"`
}
