package gzip

import (
	"net"
	"testing"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/valyala/fasthttp"

	http_context "github.com/eolinker/apinto/node/http-context"
)

var ctx http_service.IHttpContext

func getContext(address string) (http_service.IHttpContext, error) {
	if ctx == nil {
		return initTestContext(address)
	}
	if address == ctx.Request().RemoteAddr() {
		return ctx, nil
	}
	return initTestContext(address)
}
func initTestContext(address string) (http_service.IHttpContext, error) {
	fast := &fasthttp.RequestCtx{}
	freq := fasthttp.AcquireRequest()
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	fast.Init(freq, addr, nil)
	return http_context.NewContext(fast, 0), nil
}

func TestFilter(t *testing.T) {
	httpCtx, err := getContext("127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	f := NewFactory()
	d, err := f.Create("plugin@setting", "ip_restriction", "ip_restriction", "service", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	g, err := d.Create("gzip@plugin", "gzip", &Config{Types: nil, MinLength: 10, Vary: true}, nil)
	if err != nil {
		t.Errorf("create handler error : %v", err)
	}
	h, ok := g.(http_service.HttpFilter)
	if !ok {
		t.Errorf("parse filter error")
		return
	}

	cases := []struct {
		name         string
		header       string
		body         string
		wantCompress bool
	}{
		{
			name:         "wantCompress",
			wantCompress: true,
			body:         "eolink;apinto;apinto;test;gzip;eolink;apinto;apinto;test;gzip;eolink;apinto;apinto;test;gzip;eolink;apinto;apinto;test;gzip;eolink;apinto;apinto;test;gzip;eolink;apinto;apinto;test;gzip;",
			header:       "gzip",
		},
		{
			name:         "notCompress",
			wantCompress: false,
			body:         "eolink",
			header:       "",
		},
	}
	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctx.Response().SetBody([]byte(cc.body))
			ctx.Request().Header().Headers().Set("Accept-Encoding", cc.header)
			before := ctx.Response().BodyLen()
			h.DoHttpFilter(httpCtx, nil)
			after := ctx.Response().BodyLen()
			if cc.wantCompress && before <= after {
				t.Errorf("want compress; before %d, after %d", before, after)
			}
			if !cc.wantCompress && before != after {
				t.Errorf("do not want compress; before %d, after %d", before, after)
			}
		})
	}
}
