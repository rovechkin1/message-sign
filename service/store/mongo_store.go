package store

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rovechkin1/message-sign/service/config"
)

const (
	dbName             = "msg-signer"
	unsignedCollection = "records"
	signedCollection   = "signedrecords"
	signingKeys        = "signingkeys"
)

type MongoClient struct {
	Client *mongo.Client
	cancel context.CancelFunc
}
var randGenerator *rand.Rand

func init() {
	s := rand.NewSource(time.Now().UnixNano())
	randGenerator = rand.New(s)
}

func NewMongoClient(ctx context.Context) (*MongoClient, context.Context, error) {
	client, ctx, cancel, err := connect(ctx, config.GetMongoUrl())
	if err != nil {
		return nil, nil, err
	}
	// create collections if needed
	cols := make(map[string]interface{})
	db := client.Database(dbName)
	names, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, nil, err
	}

	for _, c := range names {
		cols[c] = true
	}

	need := []string{unsignedCollection, signedCollection, signingKeys}
	for _, c := range need {
		if _, ok := cols[c]; !ok {
			err := db.CreateCollection(ctx, c)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return &MongoClient{
		Client: client,
		cancel: cancel,
	}, ctx, nil
}
func (c *MongoClient) Close(ctx context.Context) {
	close(c.Client, ctx, c.cancel)
}

func (c *MongoClient) GetMongo() *mongo.Client {
	return c.Client
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

// GetRecordCount records in store which are signed
func (c *mongoStore) GetRecordCount(ctx context.Context, signed bool) (int, error) {

	db := c.client.Client.Database(dbName)
	var coll *mongo.Collection
	if !signed {
		coll = db.Collection(unsignedCollection)
	} else {
		coll = db.Collection(signedCollection)
	}

	nRecords, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	return int(nRecords), nil
}

// ReadBatch reads messages in batch
// For batch selection use sharding such that nRecords%batchCount == batchId
func (c *mongoStore) ReadBatch(ctx context.Context,
	batchId int, batchCount int) ([]Record, error) {

	// Scan all the records and select records which belong to
	// this shard. This is not efficient, e.g. each shard has to read all
	// records. Better approach would be to craft a filter query
	// to run in record-generator store
	db := c.client.Client.Database(dbName)
	coll := db.Collection(unsignedCollection)
	opts := options.Find()
	opts.SetSort(bson.D{{"id", 1}})
	sortCursor, err := coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	var records []Record
	for sortCursor.Next(ctx) == true {
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
		idBytes, err := hex.DecodeString(nr.Id)
		if err != nil {
			log.Printf("WARN: failed to convert record id : %v, skip the record", nr.Id)
			continue
		}
		if len(idBytes) < 8 {
			log.Printf("WARN: failed to convert record id, it is less than 8 bytes : %v, skip the record", nr.Id)
			continue
		}
		i := uint64(binary.LittleEndian.Uint64(idBytes[:8]))
		// check if this records belongs to batchId shard
		mod := i % uint64(batchCount)
		if mod == uint64(batchId) {
			records = append(records, nr)
		}
	}
	return records, nil
}

// WriteRecord writes a single record
func (c *mongoStore) WriteRecord(ctx context.Context, record Record) error {

	db := c.client.Client.Database(dbName)
	coll := db.Collection(unsignedCollection)
	collSign := db.Collection(signedCollection)

	filter := bson.D{{"id", record.Id}}
	update := bson.D{{"$set", bson.D{
		{"msg", record.Msg},
		{"key", record.KeyId},
		{"sign", record.Signature},
		{"salt", record.Salt},
	}}}
	opts := options.UpdateOptions{}
	opts.SetUpsert(true)
	_, err := collSign.UpdateOne(ctx, filter, update, &opts)
	if err != nil {
		return err
	}

	// record is saved, can remove it from usigned collection
	_, err = coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	log.Printf("Updated document with id %v\n", record.Id)
	return nil
}

// WriteBatch writes records as a batch
func (c *mongoStore) WriteBatch(ctx context.Context, records []Record) error {
	// TODO: check that number of written records match passed records
	db := c.client.Client.Database(dbName)
	coll := db.Collection(unsignedCollection)
	collSign := db.Collection(signedCollection)

	var docs []interface{}
	var deleteIds []string
	for _, record := range records {
		doc := bson.D{
			{"id", record.Id},
			{"msg", record.Msg},
			{"key", record.KeyId},
			{"sign", record.Signature},
			{"salt", record.Salt},
		}
		docs = append(docs, doc)
		deleteIds = append(deleteIds, record.Id)
	}

	ins, err := collSign.InsertMany(ctx, docs)
	if err != nil {
		log.Printf("ERROR: WriteBatch: Failed InsertMany, error: %v", err)
		return err
	}
	log.Printf("INFO: WriteBatch: InsertMany ok, inserted: %v", len(ins.InsertedIDs))

	// simulate test failure
	testFailureRatePct := config.GetTestSignFailureRatePct()
	if randGenerator.Intn(100) < testFailureRatePct {
		return fmt.Errorf("ERROR: This is a test failure")
	}

	// record is saved, can remove it from usigned collection
	filter := bson.M{"id": bson.M{"$in": deleteIds}}
	res, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf("ERROR: WriteBatch: Failed DeleteMany, error: %v", err)
		return err
	}
	log.Printf("INFO: WriteBatch: DeleteMany ok, deleted: %v", res)

	log.Printf("INFO: WriteBatch: Updated documents total: %v\n", len(records))

	return nil
}

// ReadKeyMetadata reads metadata of signing key
func (c *mongoStore) ReadSigningKeyMetadata(ctx context.Context, keyId string) (*SigningKeyMetadata, error) {
	db := c.client.Client.Database(dbName)
	coll := db.Collection(signingKeys)
	var result bson.D
	err := coll.FindOne(ctx, bson.D{{"id", keyId}}).Decode(&result)
	if err != nil {
		return nil, err
	}
	metadata := &SigningKeyMetadata{}
	for _, r := range result {
		switch {
		case r.Key == "id":
			metadata.Id = fmt.Sprintf("%s", r.Value)
		case r.Key == "nonce":
			metadata.Nonce = r.Value.(int64)
		}
	}
	return metadata, nil
}

// WriteSigningKeyMetadata upserts key metadata to mongo store
func (c *mongoStore) WriteSigningKeyMetadata(ctx context.Context, keyMetadata *SigningKeyMetadata) error {
	db := c.client.Client.Database(dbName)
	coll := db.Collection(signingKeys)

	filter := bson.D{{"id", keyMetadata.Id}}
	update := bson.D{{"$set", bson.D{{"nonce", keyMetadata.Nonce}}}}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func connect(ctx context.Context, uri string) (*mongo.Client, context.Context,
	context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.

	ctx, cancel := context.WithTimeout(ctx,
		30*time.Second)

	opts := options.Client().ApplyURI(uri)
	if config.GetMongoUser() != "" {
		credential := options.Credential{
			Username: config.GetMongoUser(),
			Password: config.GetMongoPwd(),
		}
		opts = opts.SetAuth(credential)
	}

	// record-generator.Connect return record-generator.Client method
	client, err := mongo.Connect(ctx, opts)
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
