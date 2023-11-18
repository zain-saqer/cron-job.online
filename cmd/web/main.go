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
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	address         = os.Getenv(`ADDRESS`)
	env             = os.Getenv(`ENV`)
	mongoHost       = os.Getenv(`MONGO_HOST`)
	mongoUsername   = os.Getenv(`MONGO_USERNAME`)
	mongoPassword   = os.Getenv(`MONGO_PASSWORD`)
	mongoPort       = os.Getenv(`MONGO_PORT`)
	mongoDatabase   = os.Getenv(`MONGO_DATABASE`)
	mongoCollection = os.Getenv(`MONGO_COLLECTION`)

	authUser = os.Getenv(`AUTH_USER`)
	authPass = os.Getenv(`AUTH_PASS`)
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, err := mongodb.NewClient(ctx, mongoHost, mongoPort, mongoUsername, mongoPassword, 3*time.Second)
	if err != nil {
		log.Fatal().Err(err).Stack().Msg(`error creating mongodb client`)
	}
	cronJobRepository := mongodb.NewMongoCronJobRepository(client, mongoDatabase, mongoCollection)
	app := App{CronJobRepository: cronJobRepository, UUIDGenerator: UUIDGenerator{}}
	e := echo.New()
	e.Debug = true
	middlewares(e, authUser, authPass)
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
		if env == `production` {
			// e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("<DOMAIN>")
			// Cache certificates to avoid issues with rate limits (https://letsencrypt.org/docs/rate-limits)
			e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
			err := e.StartAutoTLS(address)
			if err != nil && !errors.Is(http.ErrServerClosed, err) {
				log.Fatal().Err(err).Msg(`shutting down server`)
			}
		} else {
			err := e.Start(address)
			if err != nil && !errors.Is(http.ErrServerClosed, err) {
				log.Fatal().Err(err).Msg(`shutting down server`)
			}
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
