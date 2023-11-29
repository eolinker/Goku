package rate_limiting

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/drivers"
)

const (
	rateSecondType = "Second"
	rateMinuteType = "Minute"
	rateHourType   = "Hour"
	rateDayType    = "Day"
)

var _ http_service.HttpFilter = (*RateLimiting)(nil)
var _ eocontext.IFilter = (*RateLimiting)(nil)

type RateLimiting struct {
	drivers.WorkerBase
	rateInfo         *rateInfo
	hideClientHeader bool
	responseType     string
}

func (r *RateLimiting) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(r, ctx, next)
}

func (r *RateLimiting) doLimit() (bool, string, int) {
	info := r.rateInfo
	if info == nil {
		return true, "", 200
	}
	if info.second != nil {
		ok := info.second.check()
		if !ok {
			return false, "[rate_limiting] API rate limit of second exceeded", 429
		}
	}
	if info.minute != nil {
		ok := info.minute.check()
		if !ok {
			return false, "[rate_limiting] API rate limit of minute exceeded", 429
		}
	}
	if info.hour != nil {
		ok := info.hour.check()
		if !ok {
			return false, "[rate_limiting] API rate limit of hour exceeded", 429
		}
	}
	if info.day != nil {
		ok := info.day.check()
		if !ok {
			return false, "[rate_limiting] API rate limit of day exceeded", 429
		}
	}
	return true, "", 200
}

func (r *RateLimiting) Destroy() {
	r.responseType = ""
	r.rateInfo.close()
	r.rateInfo = nil
}

func (r *RateLimiting) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	// 前置处理
	flag, result, status := r.doLimit()
	if !flag {
		// 超过限制
		resp := ctx.Response()
		result = r.responseEncode(result, status)
		resp.SetStatus(403, "403")
		resp.SetBody([]byte(result))
		return err
	}
	// 后置处理
	if next != nil {
		err = next.DoChain(ctx)
	}
	if !r.hideClientHeader {
		r.addRateHeader(ctx, rateSecondType)
		r.addRateHeader(ctx, rateMinuteType)
		r.addRateHeader(ctx, rateHourType)
		r.addRateHeader(ctx, rateDayType)
	}
	return err
}

func (r *RateLimiting) Start() error {
	return nil
}

func (r *RateLimiting) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	confObj, err := drivers.Assert[Config](conf)
	if err != nil {
		return err
	}
	r.rateInfo = CreateRateInfo(confObj)
	r.hideClientHeader = confObj.HideClientHeader
	r.responseType = confObj.ResponseType
	return nil
}

func (r *RateLimiting) Stop() error {
	return nil
}

func (r *RateLimiting) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func (r *RateLimiting) responseEncode(origin string, statusCode int) string {
	if r.responseType == "json" {
		tmp := map[string]interface{}{
			"message":     origin,
			"status_code": statusCode,
		}
		newInfo, _ := json.Marshal(tmp)
		return string(newInfo)
	}
	return origin
}

func (r *RateLimiting) addRateHeader(ctx http_service.IHttpContext, rateType string) {
	var rate *rateTimer
	switch rateType {
	case rateSecondType:
		rate = r.rateInfo.second
	case rateMinuteType:
		rate = r.rateInfo.minute
	case rateHourType:
		rate = r.rateInfo.hour
	case rateDayType:
		rate = r.rateInfo.day
	}
	// 不限制
	if rate == nil || rate.limitCount == 0 || rate.requestCount == 0 {
		return
	}
	resp := ctx.Response()
	resp.SetHeader(fmt.Sprintf("X-RateLimit-Limit-%s", rateType), strconv.FormatInt(rate.limitCount, 10))
	resp.SetHeader(fmt.Sprintf("X-RateLimit-Remaining-%s", rateType), strconv.FormatInt(rate.limitCount-rate.requestCount, 10))
	return
}
