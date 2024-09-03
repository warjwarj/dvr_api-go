package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// connection to the mongodb instance
type DBConnection struct {
	logger *zap.Logger
	client *mongo.Client // connection to the db
	uri    string        // endpoint
	dbName string        // name of the database inside mongo
	lock   sync.Mutex    // lock for the db connection
}

// constructor
func NewDBConnection(logger *zap.Logger, uri string, dbName string) (*DBConnection, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	var result bson.M
	// ping db to check the connection
	err = client.Database(dbName).RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result)
	if err != nil {
		return nil, err
	}
	logger.Info("database %v connected successfully", zap.String("dbName", dbName))
	return &DBConnection{
		logger: logger,
		client: client,
		uri:    uri,
		dbName: dbName,
	}, nil
}

// insert a record into the database
func (dbc *DBConnection) RecordMessage_ToFromDevice(fromDevice bool, msg *MessageWrapper) (*mongo.UpdateResult, error) {

	// record direction
	var directionDescriptor string
	switch fromDevice {
	// msg sent by the device
	case true:
		directionDescriptor = "from device"
	// msg sent by an API client to a device
	case false:
		directionDescriptor = "to device"
	}

	// timeout, threadsafety
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	dbc.lock.Lock()

	// cancel context and release lock on database connection when func exits
	defer func() {
		cancel()
		dbc.lock.Unlock()
	}()

	// might be worth storing this to avoid redeclaration upon each function call
	coll := dbc.client.Database(dbc.dbName).Collection("devices")

	// parse the time in the packet
	packetTime, err := getDateFromMessage(msg.message)

	// Create a new message
	newMessage := DeviceMessage_Schema{
		RecvdTime:  msg.recvdTime,
		PacketTime: packetTime,
		Message:    msg.message,
		Direction:  directionDescriptor,
	}

	var devId string
	err = getIdFromMessage(&msg.message, &devId)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse device id from message: %v", msg.message)
	}

	// Update the document for the device, adding the new message to the messages array
	filter := bson.M{"DeviceId": *&devId}
	update := bson.M{
		"$push": bson.M{
			"MsgHistory": newMessage,
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

func (dbc *DBConnection) QueryMsgHistory(devices []string, before time.Time, after time.Time) ([]bson.M, error) {
	// might be worth storing this to avoid redeclaration upon each function call
	coll := dbc.client.Database(dbc.dbName).Collection("devices")

	// query filter. Device id in devices, and packet_time between the two dates passed
	filter := bson.D{
		{"DeviceId", bson.D{
			{"$in", devices},
		}},
		{"MsgHistory", bson.D{
			{"$elemMatch", bson.D{
				{"receivedTime", bson.D{
					{"$gte", after},
					{"$lt", before},
				}},
			}},
		}},
	}

	// query using above. Exclude _id field
	cursor, err := coll.Find(context.Background(), filter, options.Find().SetProjection(bson.M{"_id": 0}))
	if err != nil {
		return nil, fmt.Errorf("error querying database: %v", err)
	}

	// iterate over the cursor returned and return docuements that match the query
	var documents []bson.M
	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			panic(err)
		}
		documents = append(documents, result)
	}
	return documents, nil
}
