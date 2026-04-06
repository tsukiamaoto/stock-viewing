package model

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Source  string      `json:"source,omitempty"`
}

// NewSuccess returns a successful response.
func NewSuccess(data interface{}) APIResponse {
	return APIResponse{Status: "success", Data: data}
}

// NewError returns an error response.
func NewError(msg string) APIResponse {
	return APIResponse{Status: "error", Message: msg}
}
