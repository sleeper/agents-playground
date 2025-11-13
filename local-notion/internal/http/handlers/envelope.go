package handlers

// Envelope standardizes API responses.
type Envelope struct {
	Data   any        `json:"data,omitempty"`
	Meta   any        `json:"meta,omitempty"`
	Errors []APIError `json:"errors,omitempty"`
}

// APIError represents a structured API error message.
type APIError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}
