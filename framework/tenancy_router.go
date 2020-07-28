package imghoard

import (
	"github.com/savsgio/atreugo/v11"
)

const DefaultRoute = "_router:default"
const TenancyHeader = "x-tenancy"

type TenancyRouter struct {
	routes map[string]atreugo.View
}

func New(defaultView atreugo.View) TenancyRouter {
	router := TenancyRouter{
		routes: make(map[string]atreugo.View),
	}
	router.routes[DefaultRoute] = defaultView
	return router
}

func (router *TenancyRouter) AddTenancy(headerContent string, viewFn atreugo.View) {
	router.routes[headerContent] = viewFn
}

func (router *TenancyRouter) Route(ctx *atreugo.RequestCtx) error {
	header := ctx.Request.Header.Peek(TenancyHeader)
	if header == nil { // Header does not exist; Assume default view
		return router.routes[DefaultRoute](ctx)
	}
	if val, ok := router.routes[string(header)]; ok {
		// Header exists and a valid case is found! run custom view
		return val(ctx)
	}
	// Header exists, but no valid case seem to exist. run default view
	return router.routes[DefaultRoute](ctx)
}
