package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/application"
)

type App struct {
	drivers.WorkerBase
}

func (a *App) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	// 判断是否是websocket
	return http_service.DoHttpFilter(a, ctx, next)
}

func (a *App) Destroy() {
}

func (a *App) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	log.Debug("auth beginning")
	err := a.auth(ctx)
	if err != nil {
		ctx.Response().SetStatus(403, "403")
		ctx.Response().SetBody([]byte(err.Error()))
		return err
	}
	if next != nil {
		err = next.DoChain(ctx)
	}
	if err != nil {
		return err
	}
	return nil
}

func (a *App) DoWebsocketFilter(ctx http_service.IWebsocketContext, next eocontext.IChain) error {
	log.Debug("auth beginning")
	err := a.auth(ctx)
	if err != nil {
		ctx.Response().SetStatus(403, "403")
		ctx.Response().SetBody([]byte(err.Error()))
		return err
	}
	if next != nil {
		err = next.DoChain(ctx)
	}
	if err != nil {
		return err
	}
	return nil
}

func anonymousAppHandler(ctx http_service.IHttpContext) (bool, error) {
	if app := appManager.AnonymousApp(); app != nil && !app.Disable() {
		setLabels(ctx, app.Labels())
		ctx.SetLabel("application_id", app.Id())
		ctx.SetLabel("application", app.Name())
		log.Debug("application name is ", app.Name())
		return true, app.Execute(ctx)
	}
	return false, nil
}

func (a *App) auth(ctx http_service.IHttpContext) error {
	log.Debug("start auth...")
	if appManager.Count() < 1 {
		_, err := anonymousAppHandler(ctx)
		return err
	}
	driver := ctx.Request().Header().GetHeader("Authorization-Type")
	filters := appManager.ListByDriver(driver)

	if len(filters) < 1 {
		filters = appManager.List()
	}
	for _, filter := range filters {
		user, ok := filter.GetUser(ctx)
		if ok {
			if user == nil {
				return errors.New("invalid user")
			}
			if user.App.Disable() {
				return fmt.Errorf("the app(%s) is disabled", user.App.Id())
			}
			if user.Expire <= time.Now().Unix() && user.Expire != 0 {
				return fmt.Errorf("%s error: %s", filter.Driver(), application.ErrTokenExpired)
			}
			setLabels(ctx, user.Labels)
			setLabels(ctx, user.App.Labels())
			ctx.SetLabel("application_id", user.App.Id())
			ctx.SetLabel("application", user.App.Name())
			ctx.SetLabel("token", user.Name)
			log.Debug("application name is ", user.App.Name())
			if user.HideCredential {
				application.HideToken(ctx, user.TokenName, user.Position)
			}
			return user.App.Execute(ctx)
		}
	}
	has, err := anonymousAppHandler(ctx)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	return errors.New("missing or invalid token")
}

func setLabels(ctx http_service.IHttpContext, labels map[string]string) {
	for k, v := range labels {
		ctx.SetLabel(k, v)
	}
}

func (a *App) Start() error {
	return nil
}

func (a *App) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (a *App) Stop() error {
	return nil
}

func (a *App) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
