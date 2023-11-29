package httpoutput

import (
	"encoding/json"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"

	http_transport "github.com/eolinker/apinto/output/http-transport"
)

type Handler struct {
	formatter eosc.IFormatter
	transport formatter.ITransport
}

func NewHandler(config *Config) (*Handler, error) {
	transport, fm, err := create(config)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		formatter: fm,
		transport: transport,
	}
	return h, nil
}

func (h *Handler) Close() error {
	h.transport.Close()
	h.transport = nil
	h.formatter = nil
	return nil
}

func (h *Handler) Output(entry eosc.IEntry) error {
	if h.formatter == nil && h.transport == nil {
		return nil
	}
	data := h.formatter.Format(entry)
	if len(data) == 0 {
		return nil
	}
	return h.transport.Write(data)
}
func (h *Handler) reset(config *Config) error {

	o := h.transport
	transport, fm, err := create(config)
	if err != nil {
		return err
	}
	h.transport = transport
	h.formatter = fm

	if o != nil {
		o.Close()
	}
	return nil
}

func create(config *Config) (formatter.ITransport, eosc.IFormatter, error) {
	cfg := &http_transport.Config{
		Method:       config.Method,
		Url:          config.Url,
		Headers:      toHeader(config.Headers),
		HandlerCount: 5, // 默认值， 以后可能会改成配置
	}
	transport, err := http_transport.CreateTransporter(cfg)
	if err != nil {
		return nil, nil, err
	}

	//创建formatter
	factory, has := formatter.GetFormatterFactory(config.Type)
	if !has {
		return nil, nil, errFormatterType
	}
	var extendCfg []byte
	if config.Type == "json" {
		extendCfg, _ = json.Marshal(config.ContentResize)
	}

	fm, err := factory.Create(config.Formatter, extendCfg)
	if err != nil {
		return nil, nil, err
	}
	return transport, fm, nil
}
