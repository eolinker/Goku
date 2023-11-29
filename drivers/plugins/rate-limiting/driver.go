package rate_limiting

import (
	"reflect"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

type Driver struct {
	profession string
	name       string
	label      string
	desc       string
	configType reflect.Type
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	l := &RateLimiting{
		WorkerBase:       drivers.Worker(id, name),
		rateInfo:         CreateRateInfo(conf),
		hideClientHeader: conf.HideClientHeader,
		responseType:     conf.ResponseType,
	}
	return l, nil
}
