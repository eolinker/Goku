package httpoutput

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/output"
	scope_manager "github.com/eolinker/apinto/scope-manager"
)

var _ output.IEntryOutput = (*HttpOutput)(nil)
var _ eosc.IWorker = (*HttpOutput)(nil)

type HttpOutput struct {
	drivers.WorkerBase
	config  *Config
	handler *Handler
	running bool
}

func (h *HttpOutput) Output(entry eosc.IEntry) error {
	hd := h.handler
	if hd != nil {
		return hd.Output(entry)
	}

	return eosc.ErrorWorkerNotRunning
}

func (h *HttpOutput) Start() error {
	hd := h.handler
	if hd != nil {
		return nil
	}
	h.running = true
	handler, err := NewHandler(h.config)
	if err != nil {
		return err
	}

	h.handler = handler
	scope_manager.Set(h.Id(), h, h.config.Scopes...)
	return nil
}

func (h *HttpOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {

	config, err := check(conf)

	if err != nil {
		return err
	}
	//if h.config != nil && !h.config.isConfUpdate(config) {
	//	return nil
	//}
	h.config = config

	if h.running {
		if h.handler == nil {
			h.handler, err = NewHandler(h.config)
			if err != nil {
				return err
			}
		}

		err = h.handler.reset(config)
		if err != nil {
			return err
		}
	}
	scope_manager.Set(h.Id(), h, h.config.Scopes...)
	return nil
}

func (h *HttpOutput) Stop() error {
	scope_manager.Del(h.Id())
	hd := h.handler
	if hd != nil {
		h.handler = nil
		hd.Close()
	}
	h.config = nil
	return nil
}

func (h *HttpOutput) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
