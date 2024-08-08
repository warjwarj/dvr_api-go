package main

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// connection to the mongodb instance
type DBConnection struct {
	client     *mongo.Client // connection to the db
	uri        string        // endpoint
	db         string        // name of the database inside mongo
	deviceColl string        // name of the collection we use to store device info
	lock       sync.Mutex    // lock for the db connection
}

// constructor
func NewDBConnection(uri string) (*DBConnection, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	var result bson.M
	// ping db to check the connection
	err = client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &DBConnection{
		client:     client,
		uri:        uri,
		db:         "dvr_api-GPS_DB",
		deviceColl: "devices",
	}, nil
}

// Device schema for modelling in mongodb
type Device_Schema struct {
	devId      string                 `bson:"devId"`
	msgHistory []DeviceMessage_Schema `bson:"msgHistory"`
}

// Device message schema for modelling in mongodb
type DeviceMessage_Schema struct {
	recvdTime time.Time `bson:"timestamp"`
	message   string    `bson:"message"`
}

func (dbc *DBConnection) RecordMessage_FromDevice(msg *Message) (*mongo.UpdateResult, error) {

	// timeout, threadsafety
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	dbc.lock.Lock()

	// cancel context and release lock on database connection
	defer func() {
		cancel()
		dbc.lock.Unlock()
	}()

	// might be worth storing this to avoid redeclaration upon each function call
	coll := dbc.client.Database(dbc.db).Collection(dbc.deviceColl)

	// Create a new message
	newMessage := DeviceMessage_Schema{
		recvdTime: time.Now(),
		message:   msg.message,
	}

	// Update the document for the device, adding the new message to the messages array
	filter := bson.M{"devId": msg.id}
	update := bson.M{
		"$push": bson.M{
			"msgHistory": newMessage,
		},
	}

	// perform update/insertion
	updateResult, err := coll.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return updateResult, err
	}

	// return results of update/insertion
	return updateResult, nil
}
