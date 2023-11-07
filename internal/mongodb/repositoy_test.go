package mongodb

import (
	"context"
	"github.com/google/uuid"
	"github.com/zain-saqer/crone-job/internal/cronjob"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"
)

var (
	testDatabaseName      = `text_db`
	testCronJobCollection = `cron_job`
)

func setup(t *testing.T, client *mongo.Client) {
	t.Helper()
	ctx := context.TODO()
	err := client.Database(testDatabaseName).Drop(ctx)
	if err != nil {
		t.Error(err)
	}
	err = PrepareDatabase(ctx, client, testDatabaseName, testCronJobCollection)
	if err != nil {
		t.Error(err)
	}
}

func TestMongoCronJobRepository(t *testing.T) {
	client, err := NewClient(context.TODO(), host, port, username, password, 3*time.Second)
	if err != nil {
		t.Error(err)
	}
	setup(t, client)
	var repository cronjob.Repository = NewMongoCronJobRepository(client, testDatabaseName, testCronJobCollection)
	t1 := time.Date(2023, time.November, 10, 0, 0, 0, 0, time.Local)
	t2 := time.Date(2023, time.November, 10, 0, 1, 0, 0, time.Local)
	job1 := &cronjob.CronJob{ID: uuid.New(), NextRun: t1, CroneExpr: `5 4 * * *`, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	job2 := &cronjob.CronJob{ID: uuid.New(), NextRun: t2, CroneExpr: `5 4 * * *`, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	t.Run(`InsertCronJob and FindAllCronJobsBetween works`, func(t *testing.T) {
		ctx := context.TODO()
		_, err := repository.InsertCronJob(ctx, job1)
		if err != nil {
			t.Error(err)
		}
		_, err = repository.InsertCronJob(ctx, job2)
		if err != nil {
			t.Error(err)
		}
		jobs, err := repository.FindAllCronJobsBetween(ctx, t1, t2)
		if len(jobs) != 1 {
			t.Errorf(`unexpected len(jobs): wanted %d, got %d`, 1, len(jobs))
		}
		if job1.ID.String() != jobs[0].ID.String() {
			t.Errorf(`FindAllCronJobsBetween failed: unexpected job returned`)
		}
	})
}
