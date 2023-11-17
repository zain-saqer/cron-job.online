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
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository interface {
	FindCronJobsBetween(ctx context.Context, start, end time.Time) (<-chan CronJob, error)
	FindAllCronJobsBetween(ctx context.Context, start, end time.Time) ([]CronJob, error)
	InsertCronJob(ctx context.Context, job *CronJob) error
	UpdateOrInsert(ctx context.Context, job *CronJob) error
}

type UUIDGenerator interface {
	NewRandom() (uuid.UUID, error)
}

func New(cronExpr string, url string, now time.Time, uuidGenerator UUIDGenerator) (*CronJob, error) {
	expression, err := cronexpr.Parse(cronExpr)
	if err != nil {
		return &CronJob{}, stacktrace.Propagate(err, `invalid cron expression: %s`, cronExpr)
	}
	id, err := uuidGenerator.NewRandom()
	if err != nil {
		return &CronJob{}, stacktrace.Propagate(err, `error while generating uuid`)
	}
	return &CronJob{
		ID:        id,
		URL:       url,
		NextRun:   expression.Next(now),
		CronExpr:  cronExpr,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
