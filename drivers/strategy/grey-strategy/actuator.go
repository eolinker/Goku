package grey_strategy

import (
	"sort"
	"sync"
	"time"

	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/strategy"
)

var (
	actuatorSet ActuatorSet
)

const cookieName = "grey-cookie-%s"

func init() {
	actuator := newtActuator()
	actuatorSet = actuator
}

type ActuatorSet interface {
	strategy.IStrategyHandler
	Set(string, *GreyHandler)
	Del(id string)
}

type tActuator struct {
	lock     sync.RWMutex
	all      map[string]IGreyHandler
	handlers []IGreyHandler
}

func (a *tActuator) Destroy() {

}

func (a *tActuator) Set(id string, val *GreyHandler) {
	// 调用来源有锁
	a.all[id] = val
	a.rebuild()

}

func (a *tActuator) Del(id string) {
	// 调用来源有锁
	delete(a.all, id)
	a.rebuild()
}

func (a *tActuator) rebuild() {

	handlers := make([]IGreyHandler, 0, len(a.all))
	for _, h := range a.all {
		if !h.IsStop() {
			handlers = append(handlers, h)
		}
	}
	sort.Sort(handlerListSort(handlers))
	a.lock.Lock()
	defer a.lock.Unlock()
	a.handlers = handlers
}
func newtActuator() *tActuator {
	return &tActuator{
		all: make(map[string]IGreyHandler),
	}
}

func (a *tActuator) Strategy(ctx eocontext.EoContext, next eocontext.IChain) error {

	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		if next != nil {
			return next.DoChain(ctx)
		}
		return err
	}

	a.lock.RLock()
	handlers := a.handlers
	a.lock.RUnlock()

	for _, handler := range handlers {
		//check筛选条件
		if handler.Check(httpCtx) {
			if handler.Match(ctx) { //是否触发灰度
				handler.DoGrey(ctx)
				break
			}
		}
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

type GreyApp struct {
	scheme  string
	timeout time.Duration
	eocontext.BalanceHandler
}

func (g *GreyApp) Scheme() string {
	return g.scheme
}

func (g *GreyApp) TimeOut() time.Duration {
	return g.timeout
}

func NewGreyApp(old eocontext.BalanceHandler, grey eocontext.BalanceHandler) eocontext.BalanceHandler {
	return &GreyApp{scheme: old.Scheme(), timeout: old.TimeOut(), BalanceHandler: grey}
}

type handlerListSort []IGreyHandler

func (hs handlerListSort) Len() int {
	return len(hs)
}

func (hs handlerListSort) Less(i, j int) bool {

	return hs[i].Priority() < hs[j].Priority()
}

func (hs handlerListSort) Swap(i, j int) {
	hs[i], hs[j] = hs[j], hs[i]
}

func DoStrategy(ctx eocontext.EoContext, next eocontext.IChain) error {
	return actuatorSet.Strategy(ctx, next)
}
