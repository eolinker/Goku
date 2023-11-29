package certs

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

// Create 创建驱动实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	worker := &Worker{
		WorkerBase: drivers.Worker(id, name),
	}

	err := worker.Reset(v, workers)
	if err != nil {
		return nil, err
	}

	controller.Store(id)

	return worker, nil
}
