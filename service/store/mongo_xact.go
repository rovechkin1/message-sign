package store

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"log"
)

type MongoXact struct {
	session mongo.Session
	txnOpts *options.TransactionOptions
}

func NewMongoXact(c *mongo.Client) (*MongoXact, error) {
	var err error
	xact := &MongoXact{}
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	xact.txnOpts = options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	xact.session, err = c.StartSession()
	if err != nil {
		log.Printf("ERROR: WriteBatch: Failed StartSettion, error: %v", err)
		return nil, err
	}
	return xact, nil
}

func (c *MongoXact) WithTransaction(ctx context.Context, callback func(sessionContext mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	r, err := c.session.WithTransaction(ctx, callback, c.txnOpts)
	if err != nil {
		log.Printf("ERROR: WriteBatch: Failed WithTransaction, error: %v", err)
		return nil, err
	}
	return r, nil
}

func (c *MongoXact) Close(ctx context.Context) {
	c.session.EndSession(ctx)
}
