package driver

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
)

func Connect(m *v1.Mongo) (*mongo.Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	opts := options.Client()
	if len(m.Username) > 0 || len(m.Password) > 0 {
		opts = opts.ApplyURI(fmt.Sprintf("%s:%s@mongodb://%s:%d", m.Username, m.Password, m.Host, m.Port))
	} else {
		opts = opts.ApplyURI(fmt.Sprintf("mongodb://%s:%d", m.Host, m.Port))

	}
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	client, err := mongo.Connect(ctx, opts)
	return client, err
}

func Ping(m *v1.Mongo) error {
	client, err := Connect(m)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	return client.Ping(ctx, readpref.Primary())
}
