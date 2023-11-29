package limiting_strategy

import (
	"reflect"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

var (
	_ eosc.IWorker        = (*Limiting)(nil)
	_ eosc.IWorkerDestroy = (*Limiting)(nil)
)

type Limiting struct {
	drivers.WorkerBase
	handler   *LimitingHandler
	config    *Config
	isRunning int
}

func (l *Limiting) Destroy() error {
	controller.Del(l.Id())
	return nil
}

func (l *Limiting) Start() error {
	if l.isRunning == 0 {
		l.isRunning = 1
		actuatorSet.Set(l.Id(), l.handler)
	}

	return nil
}

func (l *Limiting) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, ok := v.(*Config)
	if !ok {
		return eosc.ErrorConfigIsNil
	}
	//if conf.Priority > 999 || conf.Priority < 1 {
	//	return fmt.Errorf("priority value %d not allow ", conf.Priority)
	//}
	confCore := conf
	if reflect.DeepEqual(l.config, confCore) {
		return nil
	}
	handler, err := NewLimitingHandler(l.Name(), confCore)
	if err != nil {
		return err
	}
	l.config = confCore
	l.handler = handler
	if l.isRunning != 0 {
		actuatorSet.Set(l.Id(), l.handler)
	}
	return nil
}

func (l *Limiting) Stop() error {
	if l.isRunning != 0 {
		l.isRunning = 0
		actuatorSet.Del(l.Id())
	}

	return nil
}

func (l *Limiting) CheckSkill(skill string) bool {
	return false
}
