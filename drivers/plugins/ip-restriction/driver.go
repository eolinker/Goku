package ip_restriction

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return v.doCheck()
}
func check(v interface{}) (*Config, error) {
	conf, err := drivers.Assert[Config](v)
	if err != nil {
		return nil, err
	}
	err = conf.doCheck()
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	err := conf.doCheck()
	if err != nil {
		return nil, err
	}
	h := &IPHandler{
		WorkerBase: drivers.Worker(id, name),
		filter:     conf.genFilter(),
	}
	return h, nil
}
