package proxy_rewrite

import (
	"fmt"
	"regexp"

	"github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.HttpFilter = (*ProxyRewrite)(nil)
var _ eocontext.IFilter = (*ProxyRewrite)(nil)

type ProxyRewrite struct {
	drivers.WorkerBase
	uri        string
	regexURI   []string
	regexMatch *regexp.Regexp
	host       string
	headers    map[string]string
}

func (p *ProxyRewrite) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(p, ctx, next)
}
func (p *ProxyRewrite) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	err = p.rewrite(ctx)
	if err != nil {
		return err
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (p *ProxyRewrite) rewrite(ctx http_service.IHttpContext) error {
	//修改uri   uri比regexURI更优先
	if p.uri != "" {
		ctx.Proxy().URI().SetPath(p.uri)
	} else if p.regexMatch != nil {
		requestURI := ctx.Proxy().URI().Path()
		if p.regexMatch.MatchString(requestURI) {
			newURI := p.regexMatch.ReplaceAllString(requestURI, p.regexURI[1])
			ctx.Proxy().URI().SetPath(newURI)
		}
	}

	//修改header中的host
	if p.host != "" {
		ctx.SetUpstreamHostHandler(upstreamHostRewrite(p.host))
	}

	//修改转发至上游的header，v可设置为空字符串，此时代表删掉header中对应的key. 若header某个key已存在则重写
	for k, v := range p.headers {
		if v == "" {
			ctx.Proxy().Header().DelHeader(k)
			continue
		}
		ctx.Proxy().Header().SetHeader(k, v)
	}

	return nil
}

func (p *ProxyRewrite) Start() error {
	return nil
}

func (p *ProxyRewrite) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}

	p.uri = conf.URI
	p.regexURI = conf.RegexURI
	p.host = conf.Host
	p.headers = conf.Headers

	if len(conf.RegexURI) > 0 {
		p.regexMatch, err = regexp.Compile(conf.RegexURI[0])
		if err != nil {
			return fmt.Errorf(regexpErrInfo, conf.RegexURI[0])
		}
	} else {
		p.regexMatch = nil
	}

	return nil
}

func (p *ProxyRewrite) Stop() error {
	return nil
}

func (p *ProxyRewrite) Destroy() {
	p.uri = ""
	p.regexURI = nil
	p.regexMatch = nil
	p.host = ""
	p.headers = nil
}

func (p *ProxyRewrite) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
