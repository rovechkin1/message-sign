package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rovechkin1/message-sign/service/config"
)

type MongoClient struct {
	Client *mongo.Client
	cancel context.CancelFunc
}

func NewMongoClient(ctx context.Context) (*MongoClient, context.Context, error) {
	client, ctx, cancel, err := connect(ctx, config.GetMongoUrl())
	if err != nil {
		return nil, nil, err
	}
	return &MongoClient{
		Client: client,
		cancel: cancel,
	}, ctx, nil
}
func (c *MongoClient) Close(ctx context.Context) {
	close(c.Client, ctx, c.cancel)
}

// Interface to read/write messages
type mongoStore struct {
	client *MongoClient
}

func NewMongoStore(client *MongoClient) MessageStore {
	return &mongoStore{
		client: client,
	}
}

// GetTotalRecords records in store
func (c *mongoStore) GetTotalRecords(ctx context.Context) (int, error) {

	db := c.client.Client.Database("msg-signer")
	coll := db.Collection("records")

	nRecords, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	return int(nRecords), nil
}

// GetTotalSignedRecords records in store which are signed
func (c *mongoStore) GetTotalSignedRecords(ctx context.Context) (int, error) {

	db := c.client.Client.Database("msg-signer")
	coll := db.Collection("records")

	nRecords, err := coll.CountDocuments(ctx, bson.D{{"sign", bson.D{{"$ne", ""}}}})
	if err != nil {
		return 0, err
	}

	return int(nRecords), nil
}

// ReadBatch reads messages in batch
func (c *mongoStore) ReadBatch(ctx context.Context,
	start int, nRecords int) ([]Record, error) {

	db := c.client.Client.Database("msg-signer")
	coll := db.Collection("records")
	opts := options.Find()
	opts.SetSort(bson.D{{"id", 1}}).SetSkip(int64(start))
	sortCursor, err := coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	var records []Record
	for i := 0; i < nRecords &&
		sortCursor.Next(ctx) == true; i += 1 {
		var result bson.D
		if err := sortCursor.Decode(&result); err != nil {
			return nil, err
		}

		nr := Record{}
		for _, r := range result {
			switch {
			case r.Key == "id":
				nr.Id = fmt.Sprintf("%s", r.Value)
			case r.Key == "msg":
				nr.Msg = fmt.Sprintf("%s", r.Value)
			case r.Key == "sign":
				nr.Signature = fmt.Sprintf("%s", r.Value)
			case r.Key == "key":
				nr.KeyId = fmt.Sprintf("%s", r.Value)
			}
		}
		records = append(records, nr)
	}
	return records, nil
}

// WriteSignaturesBatch writes messages signatures in batch
func (c *mongoStore) WriteSignaturesBatch(ctx context.Context, records []Record) error {

	db := c.client.Client.Database("msg-signer")
	coll := db.Collection("records")

	for _, r := range records {
		filter := bson.D{{"id", r.Id}}
		update := bson.D{{"$set", bson.D{{"sign", r.Signature},
			{"key", r.KeyId}}}}
		_, err := coll.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func connect(ctx context.Context, uri string) (*mongo.Client, context.Context,
	context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.

	ctx, cancel := context.WithTimeout(ctx,
		30*time.Second)

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

// This is a user defined method to close resources.
// This method closes mongoDB connection and cancel context.
func close(client *mongo.Client, ctx context.Context,
	cancel context.CancelFunc) {

	// CancelFunc to cancel to context
	defer cancel()

	// client provides a method to close
	// a mongoDB connection.
	defer func() {

		// client.Disconnect method also has deadline.
		// returns error if any,
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("ERROR: failed to disconnect, error: %v, err")
		}
	}()
}
