package http_router

import (
	http_service "github.com/eolinker/eosc/http-service"
	service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/plugin"
	router_http "github.com/eolinker/goku/router/router-http"
	service2 "github.com/eolinker/goku/service"
)

type RouterHandler struct {
	routerConfig  *router_http.Config
	routerFilters http_service.IChain
	serviceFilter service2.IService
}

func (r *RouterHandler) DoFilter(ctx service.IHttpContext, next service.IChain) (err error) {
	return r.serviceFilter.DoChain(ctx)
}

func (r *RouterHandler) Destroy() {
	s := r.serviceFilter
	if s != nil {
		r.serviceFilter = nil
		s.Destroy()
	}
	rh := r.routerFilters
	if rh != nil {
		r.routerFilters = nil
		rh.Destroy()
	}
}

func NewRouterHandler(routerConfig *router_http.Config, routerPlugin plugin.IPlugin, handler service2.IService) *RouterHandler {

	r := &RouterHandler{routerConfig: routerConfig, serviceFilter: handler}

	r.routerFilters = routerPlugin.Append(r)
	routerConfig.Target = r.routerFilters
	return r
}
