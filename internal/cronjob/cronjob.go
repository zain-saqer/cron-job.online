package cronjob

import (
	"context"
	"github.com/google/uuid"
)
import "time"

type ID uuid.UUID

func (id ID) Equals(id2 ID) bool {
	return uuid.UUID(id).String() == uuid.UUID(id2).String()
}

func NewIdFromBytes(b []byte) (ID, error) {
	newUUID, err := uuid.FromBytes(b)
	return ID(newUUID), err
}

type CronJob struct {
	ID        ID        `json:"id"`
	NextRun   time.Time `json:"next_run"`
	CroneExpr string    `json:"crone_expr"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository interface {
	FindAllCronJobsBetween(ctx context.Context, start, end time.Time) ([]CronJob, error)
	InsertCronJob(ctx context.Context, job *CronJob) (interface{}, error)
}
