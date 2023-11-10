package mongodb

import (
	"context"
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
	var results chan cronjob.CronJob
	defer close(results)
	for cursor.Next(ctx) {
		var cronJob cronjob.CronJob
		err := cursor.Decode(&cronJob)
		if err != nil {
			return nil, err
		}
		select {
		case results <- cronJob:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

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
	var result []cronjob.CronJob
	err = cursor.All(ctx, &result)

	return result, err
}

func (r *MongoCronJobRepository) InsertCronJob(ctx context.Context, job *cronjob.CronJob) (interface{}, error) {
	result, err := r.client.Database(r.database).Collection(r.collection).InsertOne(ctx, job)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func NewMongoCronJobRepository(client *mongo.Client, database, collection string) *MongoCronJobRepository {
	return &MongoCronJobRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}
