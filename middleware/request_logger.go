package middleware

import (
	"github.com/savsgio/atreugo/v11"
	"log"
)

func NewRequestLogger() atreugo.Middleware {
	return atreugo.Middleware(handleRequestLogging)
}

func handleRequestLogging(ctx *atreugo.RequestCtx) error {
	log.Printf("%s: %s\n", ctx.Method(), string(ctx.Path()) + printQueryArgs(ctx))
	_ = ctx.Next()
	return nil
}

func printQueryArgs(ctx *atreugo.RequestCtx) string {
	if ctx.QueryArgs().Len() > 0 {
		return "?" + ctx.QueryArgs().String()
	}
	return ""
}