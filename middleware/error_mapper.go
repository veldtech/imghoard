package middleware

import (
	imghoard "github.com/mikibot/imghoard/models"
	"github.com/savsgio/atreugo/v11"
)

type ErrorMapper struct {
	*atreugo.Middleware

}

func NewErrorMapper() atreugo.Middleware {
	return atreugo.Middleware(handleErrorMapping)
}

func handleErrorMapping(ctx *atreugo.RequestCtx) error {
	code := ctx.Response.StatusCode()
	if code >= 200 && code < 300 {
		return nil
	}

	return ctx.JSONResponse(imghoard.ErrorResponse{
		Message:   mapErrorString(code),
		RequestId: ctx.ID(),
	})
}

func mapErrorString(errorCode int) string {
	switch errorCode {
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	}

	return "Internal Server Error"
}