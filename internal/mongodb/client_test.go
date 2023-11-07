package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
	"time"
)

var (
	host     = os.Getenv(`MONGO_HOST`)
	username = os.Getenv(`MONGO_USERNAME`)
	password = os.Getenv(`MONGO_PASSWORD`)
	port     = os.Getenv(`MONGO_PORT`)
)

func TestNewClient(t *testing.T) {
	t.Run(`client connects to the local server`, func(t *testing.T) {
		client, err := NewClient(context.TODO(), host, port, username, password, 3*time.Second)
		if err != nil {
			t.Error(err)
		}
		// Send a ping to confirm a successful connection
		var result bson.M
		if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
			t.Error(err)
		}
		if err = client.Disconnect(context.TODO()); err != nil {
			t.Error(err)
		}
	})
}
