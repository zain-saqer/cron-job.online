package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// NewClient creates new client and connect to the server
func NewClient(ctx context.Context, host, port, username, password string, timeout time.Duration) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	uri := fmt.Sprintf(`mongodb://%s:%s/`, host, port)
	opts := options.Client().ApplyURI(uri).
		SetAuth(options.Credential{Username: username, Password: password}).
		SetServerAPIOptions(serverAPI).
		SetTimeout(timeout).
		SetBSONOptions(&options.BSONOptions{UseJSONStructTags: true}).
		SetRegistry(NewUUIDRegistry())
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	return client, nil
}
