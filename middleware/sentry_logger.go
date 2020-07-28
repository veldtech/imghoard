package middleware

import (
	"github.com/getsentry/sentry-go"
	"github.com/savsgio/atreugo/v11"
)

func HandleSentryEvent(ctx *atreugo.RequestCtx, err error) error {
	hub := sentry.CurrentHub()
	hub.CaptureEvent(&sentry.Event{
		Message: err.Error(),
		Request: &sentry.Request{
			QueryString: ctx.QueryArgs().String(),
			Method: string(ctx.Method()),
			URL: string(ctx.Request.Host()),
		},
	})
	return nil
}