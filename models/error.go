package imghoard

import (
	"github.com/palantir/stacktrace"
	"github.com/savsgio/atreugo/v9"
)

// ErrorResponse is the default error
type ErrorResponse struct {
	RequestId uint64
	Message string
}

func Error(ctx *atreugo.RequestCtx, code int, err error) error {
	return ErrorStr(ctx, code, stacktrace.Propagate(err, "").Error())
}
func ErrorStr(ctx *atreugo.RequestCtx, code int, err string) error {
	ctx.Response.SetBody([]byte(err))
	ctx.Response.SetStatusCode(code)
	return ctx.Next()
}

func JSON(ctx *atreugo.RequestCtx, json interface{}) error {
	err := ctx.JSONResponse(json)
	if err != nil {
		return Error(ctx, 400, err)
	}
	return ctx.Next()
}