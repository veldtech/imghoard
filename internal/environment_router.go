package internal

import "github.com/savsgio/atreugo/v7"

type EnvironmentRouter interface {
	AddEnv(header string, view *atreugo.View) error

	New(view *atreugo.View) (*EnvironmentRouter, error)

	Route(context *atreugo.RequestCtx)
}

type internalRouter struct {
	views       map[string]*atreugo.View
	defaultView string
}

func (router *internalRouter) AddEnv(header string, view *atreugo.View) error {

}
