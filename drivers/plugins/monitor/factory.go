package monitor

import (
	"sync"

	monitor_manager "github.com/eolinker/apinto/monitor-manager"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "monitor"
)

var (
	workers        eosc.IWorkers
	monitorManager monitor_manager.IManager
	once           sync.Once
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

type Factory struct {
	eosc.IExtenderDriverFactory
}

func NewFactory() *Factory {
	return &Factory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create),
	}
}

func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	once.Do(func() {
		bean.Autowired(&workers)
		bean.Autowired(&monitorManager)
	})

	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}
