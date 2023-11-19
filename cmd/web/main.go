package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/zain-saqer/crone-job/internal/cronjob"
	http2 "github.com/zain-saqer/crone-job/internal/http"
	"github.com/zain-saqer/crone-job/internal/mongodb"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func getConfigs() *Config {
	return &Config{
		Address:         os.Getenv(`ADDRESS`),
		MongoHost:       os.Getenv(`MONGO_HOST`),
		MongoUsername:   os.Getenv(`MONGO_USERNAME`),
		MongoPassword:   os.Getenv(`MONGO_PASSWORD`),
		MongoPort:       os.Getenv(`MONGO_PORT`),
		MongoDatabase:   os.Getenv(`MONGO_DATABASE`),
		MongoCollection: os.Getenv(`MONGO_COLLECTION`),
		AuthUser:        os.Getenv(`AUTH_USER`),
		AuthPass:        os.Getenv(`AUTH_PASS`),
	}
}

func main() {
	config := getConfigs()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, err := mongodb.NewClient(ctx, config.MongoHost, config.MongoPort, config.MongoUsername, config.MongoPassword, 3*time.Second)
	if err != nil {
		log.Fatal().Err(err).Stack().Msg(`error creating mongodb client`)
	}
	cronJobRepository := mongodb.NewMongoCronJobRepository(client, config.MongoDatabase, config.MongoCollection)
	app := App{CronJobRepository: cronJobRepository, UUIDGenerator: UUIDGenerator{}}
	e := echo.New()
	e.Debug = true
	middlewares(e, config)
	err = app.Routes(e)
	if err != nil {
		log.Fatal().Err(err).Msg(`error creating routes`)
	}
	jobServer := cronjob.JobService{
		Client:            http2.NewClient(30 * time.Second),
		CronJobRepository: cronJobRepository,
		NumberOfWorkers:   2,
		Interval:          60 * time.Second,
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		jobServer.StartPipeline(ctx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := e.Start(config.Address)
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			log.Fatal().Err(err).Msg(`shutting down server error`)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		fmt.Println(`shutting down...`)
		log.Info().Msg(`shutting down`)
		if err := e.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg(`error shutting down echo server`)
		}
	}()
	wg.Wait()
}
