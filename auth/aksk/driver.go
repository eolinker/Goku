package aksk

import (
	"github.com/eolinker/eosc"
	"reflect"
)

const (
	driverName = "aksk"
)

//driver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type driver struct {
	profession string
	name       string
	driver     string
	label      string
	desc       string
	configType reflect.Type
	params     map[string]string
}

//ConfigType 返回aksk鉴权驱动配置的反射类型
func (d *driver) ConfigType() reflect.Type {
	return d.configType
}

//Create 创建aksk鉴权驱动实例
func (d *driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	a := &aksk{
		id:   id,
		name: name,
	}
	err := a.Reset(v, workers)
	if err != nil {
		return nil, err
	}
	return a, nil
}
