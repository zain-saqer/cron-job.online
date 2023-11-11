package cronjob

import (
	"context"
	cronhttp "github.com/zain-saqer/crone-job/internal/http"
	"strings"
	"sync"
	"time"
)

type JobRequest struct {
	url  string
	time time.Time
}

type JobResult struct {
	error error
	body  string
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

func worker(ctx context.Context, jobStream <-chan JobRequest, client *cronhttp.SimpleClient) <-chan JobResult {
	results := make(chan JobResult)
	go func() {
		defer close(results)
		for {
			select {
			case job := <-jobStream:
				result, err := client.Request(ctx, `GET`, job.url, strings.NewReader(""))
				if err != nil {
					results <- JobResult{error: err}
					return
				}
				results <- JobResult{body: result.Body}
			case <-ctx.Done():
				results <- JobResult{error: ctx.Err()}
				return
			}
		}
	}()
	return results
}

func workers(ctx context.Context, jobStream <-chan JobRequest, numberOfWorkers int, client *cronhttp.SimpleClient) <-chan JobResult {
	workers := make([]<-chan JobResult, numberOfWorkers)
	for i := 0; i < numberOfWorkers; i++ {
		workers[i] = worker(ctx, jobStream, client)
	}
	return fanIn(ctx, workers...)
}

func ScheduleRunner(ctx context.Context, Client *cronhttp.SimpleClient, numberOfWorkers int, interval time.Duration, cronJobRepository Repository) <-chan JobResult {
	jobStream := make(chan JobRequest)
	workersResults := workers(ctx, jobStream, numberOfWorkers, Client)
	result := make(chan JobResult)
	go func() {
		defer close(result)
		for {
			select {
			case <-ctx.Done():
				result <- JobResult{error: ctx.Err()}
				return
			case workerResult, ok := <-workersResults:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					result <- JobResult{error: ctx.Err()}
					return
				case result <- JobResult{body: workerResult.body}:
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
				case cronJobStream, ok := <-cronJobStream:
					if !ok {
						break scheduleLoop
					}
					select {
					case <-ctx.Done():
						return
					case jobStream <- JobRequest{cronJobStream.URL, cronJobStream.NextRun}:
					}
				}
			}
		}
	}()

	return result
}
