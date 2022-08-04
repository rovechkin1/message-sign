package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

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
			panic(err)
		}
	}()
}

// This is a user defined method that returns mongo.Client,
// context.Context, context.CancelFunc and error.
// mongo.Client will be used for further database operation.
// context.Context will be used set deadlines for process.
// context.CancelFunc will be used to cancel context and
// resource associated with it.

func connect(uri string) (*mongo.Client, context.Context,
	context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

// This is a user defined method that accepts
// mongo.Client and context.Context
// This method used to ping the mongoDB, return error if any.
func insertRecords(client *mongo.Client, ctx context.Context) error {

	// mongo.Client has Ping to ping mongoDB, deadline of
	// the Ping method will be determined by cxt
	// Ping method return error if any occurred, then
	// the error can be handled.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	fmt.Println("connected successfully")
	db := client.Database("msg-signer")
	coll := db.Collection("records")

	for i := 0; i < 2; i += 1 {
		recs := generateRecords(100)
		var bsons []interface{}
		for _, r := range recs {
			obj := bson.D{
				{"id", r.Id},
				{"msg", r.Msg},
				{"sign", r.Signature},
				{"key", r.KeyId},
			}
			bsons = append(bsons, obj)
		}

		res, err := coll.InsertMany(ctx, bsons)
		if err != nil {
			return err
		}
		log.Printf("INFO: inserted %v records", len(res.InsertedIDs))
	}

	return nil
}

// Record describing message to sign
type Record struct {
	// message unique id
	Id string
	// message
	Msg string
	// signature
	Signature string
	// Public key id
	KeyId string
}

func generateRecords(nRecords int) []Record {
	var records []Record
	for i := 0; i < nRecords; i += 1 {
		uuidWithHyphen := uuid.New()
		uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
		r := Record{
			Id:  uuid,
			Msg: uuid, // same as guid
		}
		records = append(records, r)
	}
	return records
}

func main() {

	// Get Client, Context, CancelFunc and
	// err from connect method.
	client, ctx, cancel, err := connect("mongodb://localhost:27017")
	if err != nil {
		panic(err)
	}

	// Release resource when the main
	// function is returned.
	defer close(client, ctx, cancel)

	// Ping mongoDB with Ping method
	err = insertRecords(client, ctx)
	if err != nil {
		log.Printf("ERROR: failed to insert records, error: %v", err)
	}
}
