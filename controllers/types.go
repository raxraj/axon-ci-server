package controllers

type ErrorResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`       // Can be nil
	ErrorCode *int        `json:"error_code,omitempty"` // Optional error code for more specific error handling
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
