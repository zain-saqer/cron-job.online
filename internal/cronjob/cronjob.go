package cronjob

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorhill/cronexpr"
	"github.com/palantir/stacktrace"
	"time"
)

type CronJob struct {
	ID        uuid.UUID `json:"id"`
	NextRun   time.Time `json:"next_run"`
	CronExpr  string    `json:"crone_expr"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository interface {
	FindAllCronJobsBetween(ctx context.Context, start, end time.Time) ([]CronJob, error)
	InsertCronJob(ctx context.Context, job *CronJob) (interface{}, error)
}

type UUIDGenerator interface {
	New() (uuid.UUID, error)
}

func New(cronExpr string, now time.Time, uuidGenerator UUIDGenerator) (*CronJob, error) {
	expression, err := cronexpr.Parse(cronExpr)
	if err != nil {
		return &CronJob{}, stacktrace.Propagate(err, `invalid cron expression: %s`, cronExpr)
	}
	id, err := uuidGenerator.New()
	if err != nil {
		return &CronJob{}, stacktrace.Propagate(err, `error while generating uuid`)
	}
	return &CronJob{
		ID:        id,
		NextRun:   expression.Next(now),
		CronExpr:  cronExpr,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
