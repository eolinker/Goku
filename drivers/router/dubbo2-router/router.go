package dubbo2_router

import (
	"strings"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/router/dubbo2-router/manager"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/apinto/template"
)

type DubboRouter struct {
	id            string
	name          string
	manger        manager.IManger
	pluginManager plugin.IPluginManager
}

func (h *DubboRouter) Destroy() error {

	h.manger.Delete(h.id)
	return nil
}

func (h *DubboRouter) Id() string {
	return h.id
}

func (h *DubboRouter) Start() error {
	return nil
}

func (h *DubboRouter) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := drivers.Assert[Config](conf)
	if err != nil {
		return err
	}
	return h.reset(cfg, workers)

}

func (h *DubboRouter) reset(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	handler := &dubboHandler{
		completeHandler: manager.NewComplete(cfg.Retry, time.Duration(cfg.TimeOut)*time.Millisecond),
		finishHandler:   newFinishHandler(),
		routerName:      h.name,
		routerId:        h.id,
		serviceName:     strings.TrimSuffix(string(cfg.Service), "@service"),
		disable:         cfg.Disable,
		filters:         nil,
		retry:           cfg.Retry,
		timeout:         time.Duration(cfg.TimeOut) * time.Millisecond,
	}

	if !cfg.Disable {

		serviceWorker, has := workers[cfg.Service]
		if !has || !serviceWorker.CheckSkill(service.ServiceSkill) {
			return eosc.ErrorNotGetSillForRequire
		}

		if cfg.Plugins == nil {
			cfg.Plugins = map[string]*plugin.Config{}
		}

		var plugins eocontext.IChainPro
		if cfg.Template != "" {
			templateWorker, has := workers[cfg.Template]
			if !has || !templateWorker.CheckSkill(template.TemplateSkill) {
				return eosc.ErrorNotGetSillForRequire
			}
			tp := templateWorker.(template.ITemplate)
			plugins = tp.Create(h.id, cfg.Plugins)
		} else {
			plugins = h.pluginManager.CreateRequest(h.id, cfg.Plugins)
		}

		serviceHandler := serviceWorker.(service.IService)

		handler.service = serviceHandler
		handler.filters = plugins
	}

	appendRule := make([]manager.AppendRule, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		appendRule = append(appendRule, manager.AppendRule{
			Type:    r.Type,
			Name:    r.Name,
			Pattern: r.Value,
		})
	}
	err := h.manger.Set(h.id, cfg.Listen, cfg.ServiceName, cfg.MethodName, appendRule, handler)
	if err != nil {
		return err
	}
	return nil
}
func (h *DubboRouter) Stop() error {
	h.Destroy()
	return nil
}

func (h *DubboRouter) CheckSkill(skill string) bool {
	return false
}
