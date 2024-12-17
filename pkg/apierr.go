package pkg

// APIError describes an API error body that can be returned
type APIError struct {
	Message string `json:"message"`
	Error   error  `json:"error"`
}
