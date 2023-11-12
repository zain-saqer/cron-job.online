package mongodb

import (
	"context"
	"github.com/google/uuid"
	"github.com/zain-saqer/crone-job/internal/cronjob"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
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

func TestCronJobRepository(t *testing.T) {
	client, err := NewClient(context.TODO(), host, port, username, password, 3*time.Second)
	if err != nil {
		t.Error(err)
	}
	var repository = NewMongoCronJobRepository(client, testDatabaseName, testCronJobCollection)
	t1 := time.Date(2023, time.November, 10, 0, 0, 0, 0, time.Local)
	t2 := time.Date(2023, time.November, 10, 0, 1, 0, 0, time.Local)
	job1 := &cronjob.CronJob{ID: uuid.New(), NextRun: t1, CronExpr: `5 4 * * *`, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	job2 := &cronjob.CronJob{ID: uuid.New(), NextRun: t2, CronExpr: `5 4 * * *`, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	t.Run(`InsertCronJob and FindAllCronJobsBetween works`, func(t *testing.T) {
		ctx := context.TODO()
		setup(t, client)
		err := repository.InsertCronJob(ctx, job1)
		if err != nil {
			t.Error(err)
		}
		err = repository.InsertCronJob(ctx, job2)
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
	t.Run(`InsertCronJob and FindCronJobsBetween works`, func(t *testing.T) {
		ctx := context.TODO()
		setup(t, client)
		err := repository.InsertCronJob(ctx, job1)
		if err != nil {
			t.Error(err)
		}
		err = repository.InsertCronJob(ctx, job2)
		if err != nil {
			t.Error(err)
		}
		jobStream, err := repository.FindCronJobsBetween(ctx, t1, t2)
		jobs := make([]cronjob.CronJob, 0)
		for job := range jobStream {
			jobs = append(jobs, job)
		}
		if len(jobs) != 1 {
			t.Errorf(`unexpected len(jobs): wanted %d, got %d`, 1, len(jobs))
		}
		if job1.ID.String() != jobs[0].ID.String() {
			t.Errorf(`FindAllCronJobsBetween failed: unexpected job returned`)
		}
	})
	t.Run(`UpdateOrInsert insert and update a record`, func(t *testing.T) {
		ctx := context.TODO()
		setup(t, client)
		job := &cronjob.CronJob{ID: uuid.New(), NextRun: t1, CronExpr: `5 4 * * *`, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		err := repository.UpdateOrInsert(ctx, job)
		if err != nil {
			t.Error(err)
		}
		jobs, err := repository.FindAllCronJobsBetween(ctx, t1, t1.Add(time.Minute))
		if err != nil {
			t.Error(err)
		}
		if len(jobs) != 1 {
			t.Errorf(`unexpected len(jobs): wanted %d, got %d`, 1, len(jobs))
		}
		if reflect.DeepEqual(job, jobs[0]) {
			t.Errorf(`UpdateOrInsert failed: unexpected job returned`)
		}
		job.CronExpr = "cron expression"
		err = repository.UpdateOrInsert(ctx, job)
		if err != nil {
			t.Error(err)
		}
		jobs, err = repository.FindAllCronJobsBetween(ctx, t1, t1.Add(time.Minute))
		if err != nil {
			t.Error(err)
		}
		if len(jobs) != 1 {
			t.Errorf(`unexpected len(jobs): wanted %d, got %d`, 1, len(jobs))
		}
		if reflect.DeepEqual(job, jobs[0]) {
			t.Errorf(`UpdateOrInsert failed: cron-job not updated`)
		}
	})
}
