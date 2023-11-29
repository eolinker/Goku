package grpc_to_http

import (
	"fmt"
	"sync"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "grpc_to_http"
)

var (
	once   = sync.Once{}
	worker eosc.IWorkers
)

func Register(register eosc.IExtenderDriverRegister) {
	err := register.RegisterExtenderDriver(Name, NewFactory())
	if err != nil {
		log.Warnf("register %s %s", Name, err)
		return
	}
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, Check)
}

func Check(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if cfg.ProtobufID == "" {
		return fmt.Errorf("protobuf id is empty")
	}

	return nil
}
