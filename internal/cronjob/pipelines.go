package cronjob

import (
	"context"
	"github.com/gorhill/cronexpr"
	"github.com/rs/zerolog/log"
	cronhttp "github.com/zain-saqer/crone-job/internal/http"
	"strings"
	"sync"
	"time"
)

type JobRequest struct {
	cronJob *CronJob
}

type JobResult struct {
	error  error
	job    *CronJob
	body   string
	doneAt time.Time
}

func fanIn(ctx context.Context, channels ...<-chan JobResult) <-chan JobResult {
	var wg sync.WaitGroup
	multiplexedStream := make(chan JobResult)
	multiplex := func(c <-chan JobResult) {
		defer wg.Done()
		for i := range c {
			select {
			case <-ctx.Done():
				return
			case multiplexedStream <- i:
			}
		}
	}
	// Select from all the channels
	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}
	// Wait for all the reads to complete
	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()
	return multiplexedStream
}

func worker(ctx context.Context, jobStream <-chan JobRequest, client cronhttp.Client) <-chan JobResult {
	results := make(chan JobResult)
	go func() {
		defer close(results)
		for {
			select {
			case job, ok := <-jobStream:
				if !ok {
					return
				}
				log.Printf(`worker: %+v`, job.cronJob)
				result, err := client.Request(ctx, `GET`, job.cronJob.URL, strings.NewReader(""))
				if err != nil {
					log.Printf(`worker error: %s`, err.Error())
					results <- JobResult{error: err}
					return
				}
				results <- JobResult{body: result.Body, job: job.cronJob}
			case <-ctx.Done():
				results <- JobResult{error: ctx.Err()}
				return
			}
		}
	}()
	return results
}

func workers(ctx context.Context, jobStream <-chan JobRequest, numberOfWorkers int, client cronhttp.Client) <-chan JobResult {
	workers := make([]<-chan JobResult, numberOfWorkers)
	for i := 0; i < numberOfWorkers; i++ {
		workers[i] = worker(ctx, jobStream, client)
	}
	return fanIn(ctx, workers...)
}

func loop(ctx context.Context, Client cronhttp.Client, numberOfWorkers int, interval time.Duration, cronJobRepository Repository) <-chan JobResult {
	jobStream := make(chan JobRequest)
	workersResults := workers(ctx, jobStream, numberOfWorkers, Client)
	result := make(chan JobResult)
	go func() {
		defer close(result)
		for {
			select {
			case <-ctx.Done():
				return
			case workerResult, ok := <-workersResults:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case result <- JobResult{body: workerResult.body, job: workerResult.job}:
				}
			}
		}
	}()
	go func() {
		defer close(jobStream)
		for range time.Tick(interval) {
			now := time.Now()
			cronJobStream, err := cronJobRepository.FindCronJobsBetween(ctx, time.Unix(0, 0), now.Add(interval))
			if err != nil {
				select {
				case <-ctx.Done():
				case result <- JobResult{error: err}:
				}
				return
			}
		scheduleLoop:
			for {
				select {
				case <-ctx.Done():
					return
				case cronJob, ok := <-cronJobStream:
					if !ok {
						break scheduleLoop
					}
					select {
					case <-ctx.Done():
						return
					case jobStream <- JobRequest{cronJob: &cronJob}:
					}
				}
			}
		}
	}()

	return result
}

func processResults(ctx context.Context, resultStream <-chan JobResult, cronJobRepository Repository) {
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-resultStream:
			if result.error != nil {
				log.Printf(`result processing error: %s`, result.error.Error())
				continue
			}
			cronExpr, err := cronexpr.Parse(result.job.CronExpr)
			if err != nil {
				log.Printf(`result processing error: %s`, err.Error())
				continue
			}
			result.job.NextRun = cronExpr.Next(result.job.NextRun)
			err = cronJobRepository.UpdateOrInsert(ctx, result.job)
			if err != nil {
				log.Printf(`result processing error: %s`, err.Error())
				continue
			}
			log.Printf(`result processing: job done and updated %+v`, result.job)
		}
	}
}

type JobService struct {
	Client            cronhttp.Client
	CronJobRepository Repository
	NumberOfWorkers   int
	Interval          time.Duration
}

func (s *JobService) StartPipeline(ctx context.Context) {
	resultStream := loop(ctx, s.Client, s.NumberOfWorkers, s.Interval, s.CronJobRepository)
	processResults(ctx, resultStream, s.CronJobRepository)
}
