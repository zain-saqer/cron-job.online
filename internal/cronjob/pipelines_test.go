package cronjob

import (
	"context"
	"github.com/google/uuid"
	cronhttp "github.com/zain-saqer/crone-job/internal/http"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func Test_requestWorker(t *testing.T) {
	t.Run(`worker sends requests`, func(t *testing.T) {
		var callCount atomic.Int64
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount.Add(1)
		}))
		defer testServer.Close()
		ctx := context.TODO()
		jobStream := make(chan JobRequest)
		client := cronhttp.NewClient(time.Second)
		results := worker(ctx, jobStream, client)
		var jobNum int64 = 100
		go func() {
			for i := int64(0); i < jobNum; i++ {
				jobStream <- JobRequest{url: testServer.URL}
			}
			close(jobStream)
		}()
	LOOP:
		for {
			select {
			case _, ok := <-results:
				if !ok {
					if callCount.Load() != jobNum {
						t.Errorf(`want callCount to be %d, got %d`, jobNum, callCount.Load())
					}
					break LOOP
				}
			case <-time.After(time.Second):
				t.Errorf(`timeout`)
				break LOOP
			}
		}
	})
}

func Test_startWorkers(t *testing.T) {
	t.Run(`workers send requests`, func(t *testing.T) {
		var callCount atomic.Int64
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount.Add(1)
		}))
		defer testServer.Close()
		ctx := context.TODO()
		jobStream := make(chan JobRequest)
		client := cronhttp.NewClient(time.Second)
		results := workers(ctx, jobStream, 4, client)
		var jobNum int64 = 100
		go func() {
			for i := int64(0); i < jobNum; i++ {
				jobStream <- JobRequest{url: testServer.URL}
			}
			close(jobStream)
		}()
	LOOP:
		for {
			select {
			case _, ok := <-results:
				if !ok {
					if callCount.Load() != jobNum {
						t.Errorf(`want callCount to be %d, got %d`, jobNum, callCount.Load())
					}
					break LOOP
				}
			case <-time.After(time.Second):
				t.Errorf(`timeout`)
				break LOOP
			}
		}
	})
}

type MockRepo struct {
	results       []CronJob
	counter       atomic.Int64
	jobCountLimit int64
}

func (r *MockRepo) FindCronJobsBetween(context.Context, time.Time, time.Time) (<-chan CronJob, error) {
	result := make(chan CronJob)
	go func() {
		defer close(result)
		for _, cronJob := range r.results {
			if r.jobCountLimit == r.counter.Load() {
				return
			}
			result <- cronJob
			r.counter.Add(1)
		}
	}()
	return result, nil
}
func (r *MockRepo) FindAllCronJobsBetween(context.Context, time.Time, time.Time) ([]CronJob, error) {
	return nil, nil
}
func (r *MockRepo) InsertCronJob(context.Context, *CronJob) error {
	return nil
}
func (r *MockRepo) UpdateOrInsert(context.Context, *CronJob) error {
	return nil
}

func TestScheduleRunner(t *testing.T) {
	var callCount atomic.Int64
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
	}))
	defer testServer.Close()
	cronJobs := make([]CronJob, 0)
	for i := 0; i < 10; i++ {
		job := CronJob{ID: uuid.New(), NextRun: time.Now(), CronExpr: `5 4 * * *`, CreatedAt: time.Now(), UpdatedAt: time.Now(), URL: testServer.URL}
		cronJobs = append(cronJobs, job)
	}

	t.Run(`ScheduleRunner sends requests`, func(t *testing.T) {
		timeout := 200 * time.Millisecond
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(timeout, func() {
			cancel()
		})
		client := cronhttp.NewClient(timeout)
		var cronJobLimit int64 = 1000
		repo := &MockRepo{results: cronJobs, jobCountLimit: cronJobLimit}
		results := ScheduleRunner(ctx, client, 3, 100, repo)
	LOOP:
		for {
			select {
			case _, ok := <-results:
				if !ok {
					if callCount.Load() != repo.counter.Load() || callCount.Load() != repo.jobCountLimit {
						t.Errorf(`want callCount to be %d, got %d`, cronJobLimit, repo.counter.Load())
					}
					cancel()
					break LOOP
				}
			}
		}
	})
}
