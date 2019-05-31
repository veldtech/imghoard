package imghoard

// ErrorResponse is the default error
type ErrorResponse struct {
	Error string
}

// New creates a new error response
func New(reason string) ErrorResponse {
	return ErrorResponse {
		Error: reason,
	}
}