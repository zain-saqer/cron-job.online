package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/zain-saqer/crone-job/internal/cronjob"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoCronJobRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

func PrepareDatabase(ctx context.Context, client *mongo.Client, database, collection string) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"id", 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := client.Database(database).Collection(collection).Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoCronJobRepository) FindCronJobsBetween(ctx context.Context, start, end time.Time) (<-chan cronjob.CronJob, error) {
	filter := bson.D{
		{`next_run`, bson.D{{`$gte`, start}}},
		{`next_run`, bson.D{{`$lt`, end}}},
	}
	cursor, err := r.client.Database(r.database).Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	results := make(chan cronjob.CronJob)
	go func() {
		defer close(results)
		for cursor.Next(ctx) {
			var cronJob cronjob.CronJob
			err := cursor.Decode(&cronJob)
			if err != nil {
				return
			}
			select {
			case results <- cronJob:
				fmt.Println(`case results <- cronJob:`)
			case <-ctx.Done():
				return
			}
		}
	}()

	return results, err
}

func (r *MongoCronJobRepository) FindAllCronJobsBetween(ctx context.Context, start, end time.Time) ([]cronjob.CronJob, error) {
	filter := bson.D{
		{`next_run`, bson.D{{`$gte`, start}}},
		{`next_run`, bson.D{{`$lt`, end}}},
	}
	cursor, err := r.client.Database(r.database).Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	result := make([]cronjob.CronJob, 0)
	err = cursor.All(ctx, &result)

	return result, err
}

func (r *MongoCronJobRepository) InsertCronJob(ctx context.Context, job *cronjob.CronJob) error {
	_, err := r.client.Database(r.database).Collection(r.collection).InsertOne(ctx, job)
	if err != nil {
		return err
	}
	return nil
}

func (r *MongoCronJobRepository) UpdateOrInsert(ctx context.Context, job *cronjob.CronJob) error {
	filter := bson.D{
		{`id`, job.ID},
	}
	job.UpdatedAt = time.Now()
	result := r.client.Database(r.database).Collection(r.collection).FindOneAndReplace(ctx, filter, job)
	err := result.Err()
	if err == nil {
		return nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	return r.InsertCronJob(ctx, job)
}

func NewMongoCronJobRepository(client *mongo.Client, database, collection string) *MongoCronJobRepository {
	return &MongoCronJobRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}
