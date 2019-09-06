package imghoard

import "github.com/savsgio/atreugo/v7"

// ErrorResponse is the default error
type ErrorResponse struct {
	Error string
}

func NewJSON(reason string) atreugo.JSON {
	return atreugo.JSON {
		"error": reason,
	}
}