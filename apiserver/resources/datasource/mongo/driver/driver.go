package driver

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	mg "github.com/chinamobile/nlpt/apiserver/resources/datasource/mongo"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
)

func Connect(m *v1.Mongo) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := options.Client()
	uri := ""
	// standard mongo uri
	//mongodb://[username:password@]host1[:port1][,...hostN[:portN]]][/[database][?options]]
	if len(m.Username) > 0 || len(m.Password) > 0 {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/", m.Username, m.Password, m.Host, m.Port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/", m.Host, m.Port)
	}
	if len(m.Database) > 0 {
		uri = uri + m.Database
	}
	opts = opts.ApplyURI(uri)
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return client.Ping(ctx, readpref.Primary())
}

func FindCollections(m *v1.Mongo) ([]mg.Collection, error) {
	client, err := Connect(m)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := client.Database(m.Database).ListCollections(ctx, struct{}{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	result := make([]mg.Collection, 0)
	for cur.Next(ctx) {
		var col bson.M
		err := cur.Decode(&col)
		if err != nil {
			return nil, err
		}
		name, ok := col["name"]
		if !ok {
			return nil, fmt.Errorf("cannot find collection name")
		}
		nameStr := name.(string)
		result = append(result, mg.Collection{nameStr})
	}
	return result, nil
}

func FindFields(m *v1.Mongo, collectionName string) ([]mg.Field, error) {
	client, err := Connect(m)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := client.Database(m.Database).Collection(collectionName)

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	result := make([]mg.Field, 0)
	keys := make(map[string]bool)
	for cur.Next(ctx) {
		var doc bson.M
		err := cur.Decode(&doc)
		if err != nil {
			return nil, err
		}
		for k, v := range doc {
			if _, ok := keys[k]; ok {
				continue
			}
			keys[k] = true
			var tp mg.DataType = mg.Unknown
			switch v.(type) {
			case bool:
				tp = mg.Bool
				break
			case int, int8, int16, int32, int64:
				tp = mg.Integer
				break
			case float64:
				tp = mg.Float
				break
			case string:
				tp = mg.String
				break
			default:
			}
			result = append(result, mg.Field{Name: k, DataType: tp})
		}
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
