package redis

import (
	"reflect"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"

	"github.com/eolinker/apinto/drivers"
)

var (
	configType = reflect.TypeOf(new(Config))
	render     interface{}
)

func init() {
	render, _ = schema.Generate(configType, nil)

}

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver("redis", NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {

	return drivers.NewFactory[Config](Create)
}
