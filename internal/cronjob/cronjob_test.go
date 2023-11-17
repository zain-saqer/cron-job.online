package cronjob

import (
	"github.com/google/uuid"
	"reflect"
	"testing"
	"time"
)

type mockUUIDGenerator struct {
	id uuid.UUID
}

func (g mockUUIDGenerator) NewRandom() (uuid.UUID, error) {
	return g.id, nil
}

func TestCronJob_New(t *testing.T) {
	t.Run(`NewRandom returns expected CronJob`, func(t *testing.T) {
		now := time.Date(2023, 11, 7, 22, 20, 0, 0, time.Local)
		fiveMinLater := now.Add(5 * time.Minute)
		expectedId := uuid.New()
		cronExpression := `*/5 * * * *`
		expected := &CronJob{
			ID:        expectedId,
			NextRun:   fiveMinLater,
			CronExpr:  cronExpression,
			CreatedAt: now,
			UpdatedAt: now,
		}
		resultCronJob, err := New(cronExpression, ``, now, &mockUUIDGenerator{expectedId})
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(expected, resultCronJob) {
			t.Errorf(`unexpected CronJob`)
		}
	})
	t.Run(`NewRandom returns error when cron-expression is invalid`, func(t *testing.T) {
		now := time.Date(2023, 11, 7, 22, 20, 0, 0, time.Local)
		_, err := New(``, ``, now, &mockUUIDGenerator{uuid.New()})
		if err == nil {
			t.Errorf(`error was expected`)
		}
	})
}
