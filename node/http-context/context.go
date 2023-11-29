package http_context

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/utils/config"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/eolinker/apinto/entries/ctx_key"
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	fasthttp_client "github.com/eolinker/apinto/node/fasthttp-client"
)

var _ http_service.IHttpContext = (*HttpContext)(nil)

// HttpContext fasthttpRequestCtx
type HttpContext struct {
	fastHttpRequestCtx *fasthttp.RequestCtx
	proxyRequest       ProxyRequest
	proxyRequests      []http_service.IProxy
	requestID          string
	response           Response
	requestReader      RequestReader
	ctx                context.Context
	completeHandler    eoscContext.CompleteHandler
	finishHandler      eoscContext.FinishHandler

	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	labels              map[string]string
	port                int
	entry               eosc.IEntry
}

func (ctx *HttpContext) RealIP() string {
	return ctx.Request().RealIp()
}

func (ctx *HttpContext) GetUpstreamHostHandler() eoscContext.UpstreamHostHandler {
	return ctx.upstreamHostHandler
}

func (ctx *HttpContext) SetUpstreamHostHandler(handler eoscContext.UpstreamHostHandler) {
	ctx.upstreamHostHandler = handler
}

func (ctx *HttpContext) LocalIP() net.IP {
	return ctx.fastHttpRequestCtx.LocalIP()
}

func (ctx *HttpContext) LocalAddr() net.Addr {
	return ctx.fastHttpRequestCtx.LocalAddr()
}

func (ctx *HttpContext) LocalPort() int {
	return ctx.port
}

func (ctx *HttpContext) GetBalance() eoscContext.BalanceHandler {
	return ctx.balance
}

func (ctx *HttpContext) SetBalance(handler eoscContext.BalanceHandler) {
	ctx.balance = handler
}

func (ctx *HttpContext) GetEntry() eosc.IEntry {
	if ctx.entry == nil {
		ctx.entry = http_entry.NewEntry(ctx)
	}
	return ctx.entry
}

func (ctx *HttpContext) SetLabel(name, value string) {
	ctx.labels[name] = value
}

func (ctx *HttpContext) GetLabel(name string) string {
	return ctx.labels[name]
}

func (ctx *HttpContext) Labels() map[string]string {
	return ctx.labels
}

func (ctx *HttpContext) GetComplete() eoscContext.CompleteHandler {
	return ctx.completeHandler
}

func (ctx *HttpContext) SetCompleteHandler(handler eoscContext.CompleteHandler) {
	ctx.completeHandler = handler
}

func (ctx *HttpContext) GetFinish() eoscContext.FinishHandler {
	return ctx.finishHandler
}

func (ctx *HttpContext) SetFinish(handler eoscContext.FinishHandler) {
	ctx.finishHandler = handler
}

func (ctx *HttpContext) Scheme() string {
	return string(ctx.fastHttpRequestCtx.Request.URI().Scheme())
}

func (ctx *HttpContext) Assert(i interface{}) error {
	if v, ok := i.(*http_service.IHttpContext); ok {
		*v = ctx
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (ctx *HttpContext) Proxies() []http_service.IProxy {
	return ctx.proxyRequests
}

func (ctx *HttpContext) Response() http_service.IResponse {
	return &ctx.response
}

func (ctx *HttpContext) SendTo(scheme string, node eoscContext.INode, timeout time.Duration) error {
	host := node.Addr()
	request := ctx.proxyRequest.Request()

	passHost, targetHost := ctx.GetUpstreamHostHandler().PassHost()
	switch passHost {
	case eoscContext.PassHost:
	case eoscContext.NodeHost:
		request.URI().SetHost(host)
	case eoscContext.ReWriteHost:
		request.URI().SetHost(targetHost)
	}
	beginTime := time.Now()
	ctx.response.responseError = fasthttp_client.ProxyTimeout(scheme, node, request, &ctx.fastHttpRequestCtx.Response, timeout)
	agent := newRequestAgent(&ctx.proxyRequest, host, scheme, beginTime, time.Now())
	if ctx.response.responseError != nil {
		agent.setStatusCode(504)
	} else {
		ctx.response.ResponseHeader.refresh()
		ip, port := parseAddr(ctx.fastHttpRequestCtx.Response.RemoteAddr().String())
		agent.setRemoteIP(ip)
		agent.setRemotePort(port)
		ctx.response.remoteIP = ip
		ctx.response.remotePort = port
		agent.setStatusCode(ctx.fastHttpRequestCtx.Response.StatusCode())
	}
	agent.responseBody = string(ctx.response.Response.Body())

	agent.setResponseLength(ctx.fastHttpRequestCtx.Response.Header.ContentLength())

	ctx.proxyRequests = append(ctx.proxyRequests, agent)
	return ctx.response.responseError
}

func (ctx *HttpContext) Context() context.Context {
	if ctx.ctx == nil {
		ctx.ctx = context.Background()
	}
	return ctx.ctx
}

func (ctx *HttpContext) AcceptTime() time.Time {
	return ctx.fastHttpRequestCtx.Time()
}

func (ctx *HttpContext) Value(key interface{}) interface{} {
	return ctx.Context().Value(key)
}

func (ctx *HttpContext) WithValue(key, val interface{}) {
	ctx.ctx = context.WithValue(ctx.Context(), key, val)
}

func (ctx *HttpContext) Proxy() http_service.IRequest {
	return &ctx.proxyRequest
}

func (ctx *HttpContext) Request() http_service.IRequestReader {

	return &ctx.requestReader
}

func (ctx *HttpContext) IsCloneable() bool {
	return true
}

func (ctx *HttpContext) Clone() (eoscContext.EoContext, error) {
	copyContext := copyPool.Get().(*cloneContext)
	copyContext.org = ctx
	copyContext.proxyRequests = make([]http_service.IProxy, 0, 2)

	req := fasthttp.AcquireRequest()
	// 当body未读取，调用Body方法读出stream中当所有body内容，避免请求体被截断
	ctx.proxyRequest.req.Body()
	ctx.proxyRequest.req.CopyTo(req)

	resp := fasthttp.AcquireResponse()
	// ctx.fastHttpRequestCtx.Response.CopyTo(resp)

	copyContext.proxyRequest.reset(req, ctx.requestReader.remoteAddr)
	copyContext.response.reset(resp)

	copyContext.completeHandler = ctx.completeHandler
	copyContext.finishHandler = ctx.finishHandler

	cloneLabels := make(map[string]string, len(ctx.labels))
	for k, v := range ctx.labels {
		cloneLabels[k] = v
	}
	copyContext.labels = cloneLabels

	// 记录请求时间
	copyContext.ctx = context.WithValue(ctx.Context(), http_service.KeyCloneCtx, true)
	copyContext.WithValue(ctx_key.CtxKeyRetry, 0)
	return copyContext, nil
}

// NewContext 创建Context
func NewContext(ctx *fasthttp.RequestCtx, port int) *HttpContext {
	remoteAddr := ctx.RemoteAddr().String()
	httpContext := pool.Get().(*HttpContext)

	httpContext.fastHttpRequestCtx = ctx
	httpContext.requestID = uuid.New().String()

	// 原始请求最大读取body为8k，使用clone request
	request := fasthttp.AcquireRequest()

	if ctx.Request.IsBodyStream() && ctx.Request.Header.ContentLength() > 8*1024 {
		ctx.Request.Body()
	}
	ctx.Request.CopyTo(request)
	httpContext.requestReader.reset(request, remoteAddr)

	// proxyRequest保留原始请求
	httpContext.proxyRequest.reset(&ctx.Request, remoteAddr)
	httpContext.proxyRequests = httpContext.proxyRequests[:0]
	httpContext.response.reset(&ctx.Response)
	httpContext.labels = make(map[string]string)
	httpContext.port = port
	// 记录请求时间
	httpContext.ctx = context.Background()
	httpContext.WithValue("request_time", ctx.Time())

	return httpContext
}

// RequestId 请求ID
func (ctx *HttpContext) RequestId() string {
	return ctx.requestID
}

// Finish finish
func (ctx *HttpContext) FastFinish() {
	if ctx.response.responseError != nil {
		ctx.fastHttpRequestCtx.SetStatusCode(504)
		ctx.fastHttpRequestCtx.SetBodyString(ctx.response.responseError.Error())
		return
	}

	ctx.port = 0
	ctx.ctx = nil
	ctx.balance = nil
	ctx.upstreamHostHandler = nil
	ctx.finishHandler = nil
	ctx.completeHandler = nil
	fasthttp.ReleaseRequest(ctx.requestReader.req)

	ctx.requestReader.Finish()
	ctx.proxyRequest.Finish()
	ctx.response.Finish()
	ctx.fastHttpRequestCtx = nil
	pool.Put(ctx)
}

func parseAddr(addr string) (string, int) {
	a := strings.Split(addr, ":")
	port := 0
	if len(a) > 1 {
		port, _ = strconv.Atoi(a[1])
	}
	return a[0], port
}
