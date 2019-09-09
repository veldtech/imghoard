package imghoard

import "github.com/savsgio/atreugo/v7"

// ErrorResponse is the default error
type ErrorResponse struct {
	Error string
}

// Deprecated
func NewJSON(reason string) atreugo.JSON {
	return newJSON(500, reason)
}

func BadRequest(ctx *atreugo.RequestCtx, reason ...string) error {
	return ctx.JSONResponse(
		newJSON(400, getOrDefault("Bad request", reason...)), 400)
}

func NotFound(ctx *atreugo.RequestCtx, reason ...string) error {
	return ctx.JSONResponse(newJSON(404, getOrDefault("Not found", reason...)), 404)
}

func InternalServerError(ctx *atreugo.RequestCtx, reason ...string) error {
	return ctx.JSONResponse(newJSON(500, getOrDefault("Internal server error", reason...)), 500)
}

func newJSON(code int, reason string) atreugo.JSON {
	return atreugo.JSON{
		"code":  code,
		"error": reason,
	}
}

func getOrDefault(d string, v ...string) string {
	for _, x := range v {
		if len(x) > 0 {
			return x
		}
	}
	return d
}
