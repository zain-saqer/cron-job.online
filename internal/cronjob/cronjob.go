package cronjob

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type CronJob struct {
	ID        uuid.UUID `json:"id"`
	NextRun   time.Time `json:"next_run"`
	CroneExpr string    `json:"crone_expr"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository interface {
	FindAllCronJobsBetween(ctx context.Context, start, end time.Time) ([]CronJob, error)
	InsertCronJob(ctx context.Context, job *CronJob) (interface{}, error)
}
