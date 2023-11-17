package main

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/zain-saqer/crone-job/internal/cronjob"
	"github.com/zain-saqer/crone-job/web"
	"html/template"
	"net/http"
	"sync"
	"time"
)

func (app App) Routes(e *echo.Echo) error {
	e.GET(`/`, app.index)
	e.GET(`/cronjob/list`, app.getCronJobList)
	e.GET(`/cronjob/add`, app.getCronJobAdd)
	e.POST(`/cronjob/add`, app.postCronJobAdd)

	return nil
}

func (app App) index(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, `/cronjob/list`)
}

func (app App) getCronJobList(c echo.Context) error {
	var t *template.Template
	sync.OnceFunc(func() {
		var err error
		t, err = template.ParseFS(web.F, `templates/layout.gohtml`, `templates/nav.gohtml`, `templates/cronjob/list.gohtml`)
		if err != nil {
			log.Fatal().Err(err).Stack().Msg(`error parsing templates`)
		}
	})()
	now := time.Now()
	jobs, err := app.CronJobRepository.FindAllCronJobsBetween(c.Request().Context(), time.Unix(0, 0), now.Add(5*time.Minute))
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(c.Response(), `base`, CronjobViewData{
		CronjobList: jobs,
	})
}

func (app App) getCronJobAdd(c echo.Context) error {
	var t *template.Template
	sync.OnceFunc(func() {
		var err error
		t, err = template.ParseFS(web.F, `templates/layout.gohtml`, `templates/nav.gohtml`, `templates/cronjob/add.gohtml`)
		if err != nil {
			log.Fatal().Err(err).Stack().Msg(`error parsing templates`)
		}
	})()
	return t.ExecuteTemplate(c.Response(), `base`, CronjobAdd{})
}

func (app App) postCronJobAdd(c echo.Context) error {
	var t *template.Template
	sync.OnceFunc(func() {
		var err error
		t, err = template.ParseFS(web.F, `templates/layout.gohtml`, `templates/nav.gohtml`, `templates/cronjob/add.gohtml`)
		if err != nil {
			log.Fatal().Err(err).Stack().Msg(`error parsing templates`)
		}
	})()
	cronjobAdd := &CronjobAdd{}
	err := c.Bind(cronjobAdd)
	if err != nil {
		return err
	}
	validationErrors := cronjobAdd.Validate()
	if len(validationErrors) > 0 {
		return t.ExecuteTemplate(c.Response(), `base`, cronjobAdd)
	}
	job, err := cronjob.New(cronjobAdd.CronExpr, cronjobAdd.URL, time.Now(), app.UUIDGenerator)
	err = app.CronJobRepository.InsertCronJob(c.Request().Context(), job)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, `/cronjob/list`)
}
