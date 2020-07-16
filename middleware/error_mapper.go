package middleware

import (
	"fmt"
	imghoard "github.com/mikibot/imghoard/models"
	"github.com/savsgio/atreugo/v9"
)

type ErrorMapper struct {
	*atreugo.Middleware
}

func NewErrorMapper() atreugo.Middleware {
	view := atreugo.Middleware(handleErrorMapping);
	return view
}

func handleErrorMapping(ctx *atreugo.RequestCtx) error {
	code := ctx.Response.StatusCode()

	if code >= 200 && code < 300 {
		return nil
	}

	if code >= 200 {
		fmt.Printf("ERROR: %d\n%s", ctx.ID(), string(ctx.Response.Body()))
	}

	return ctx.JSONResponse(imghoard.ErrorResponse{
		Message: mapErrorString(code),
		RequestId: ctx.ID(),
	}, code)
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