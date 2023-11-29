package dubbo2_router

import (
	"errors"
	"time"

	"github.com/eolinker/apinto/entries/ctx_key"

	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"

	"github.com/eolinker/apinto/drivers/router/dubbo2-router/manager"
	"github.com/eolinker/apinto/router"
	"github.com/eolinker/apinto/service"
)

var _ router.IRouterHandler = (*dubboHandler)(nil)

type dubboHandler struct {
	completeHandler eocontext.CompleteHandler
	finishHandler   eocontext.FinishHandler
	routerName      string
	routerId        string
	serviceName     string
	disable         bool
	service         service.IService
	filters         eocontext.IChainPro
	retry           int
	timeout         time.Duration
	labels          map[string]string
}

var completeCaller = manager.NewCompleteCaller()

func (d *dubboHandler) ServeHTTP(ctx eocontext.EoContext) {

	dubboCtx, err := dubbo2_context.Assert(ctx)
	if err != nil {
		return
	}

	if d.disable {
		dubboCtx.Response().SetBody(manager.Dubbo2ErrorResult(errors.New("router disable")))
		return
	}
	for key, value := range d.labels {
		ctx.SetLabel(key, value)
	}

	//set retry timeout
	ctx.WithValue(ctx_key.CtxKeyRetry, d.retry)
	ctx.WithValue(ctx_key.CtxKeyTimeout, d.timeout)

	//Set Label
	ctx.SetLabel("api", d.routerName)
	ctx.SetLabel("api_id", d.routerId)
	ctx.SetLabel("service", d.serviceName)
	ctx.SetLabel("service_id", d.service.Id())
	ctx.SetLabel("ip", dubboCtx.HeaderReader().RemoteIP())

	ctx.SetCompleteHandler(d.completeHandler)
	ctx.SetFinish(d.finishHandler)
	ctx.SetBalance(d.service)
	ctx.SetUpstreamHostHandler(d.service)

	_ = d.filters.Chain(ctx, completeCaller)

}
