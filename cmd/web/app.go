package main

import (
	"github.com/google/uuid"
	"github.com/zain-saqer/crone-job/internal/cronjob"
)

type App struct {
	CronJobRepository cronjob.Repository
	UUIDGenerator     cronjob.UUIDGenerator
}

type UUIDGenerator struct {
}

func (g UUIDGenerator) NewRandom() (uuid.UUID, error) {
	return uuid.NewRandom()
}
