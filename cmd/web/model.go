package main

import (
	"github.com/gorhill/cronexpr"
	"github.com/zain-saqer/crone-job/internal/cronjob"
	"net/url"
)

type CronjobViewData struct {
	CronjobList []cronjob.CronJob
}

type CronjobAdd struct {
	URL      string `form:"url"`
	CronExpr string `form:"cronExpr"`
	Errors   []ValidationError
}

type ValidationError struct {
	Field   string
	Message string
}

func (c *CronjobAdd) Validate() []ValidationError {
	errors := make([]ValidationError, 0)
	_, err := cronexpr.Parse(c.CronExpr)
	if err != nil {
		errors = append(errors, ValidationError{`CronExpr`, `Invalid cron expression`})
	}
	_, err = url.Parse(c.URL)
	if err != nil {
		errors = append(errors, ValidationError{`URL`, `Invalid URL`})
	}
	c.Errors = errors
	return errors
}
