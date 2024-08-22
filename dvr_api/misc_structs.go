package main

import (
	"time"
)

// used in ws_svr.go to read json messages into structs like this
type JsonWSMessage struct {
	Messages      []string `json:"messages"`
	Subscriptions []string `json:"subscriptions"`
}

// might be able to get rid of this
type Message struct {
	message  string  // text the tcp client sent
	clientId *string // index which the message sender with in the connIndex of the server
}

// Device schema for modelling in mongodb
type Device_Schema struct {
	DeviceId   string                 `bson:"device_id"`
	MsgHistory []DeviceMessage_Schema `bson:"msg_history"`
}

// Device message schema for modelling in mongodb
type DeviceMessage_Schema struct {
	RecvdTime  time.Time `bson:"received_time"`
	PacketTime time.Time `bson:"packet_time"`
	Message    string    `bson:"message"`
}

// time is in the ISO string format
type API_Request struct {
	Devices []string  `bson:"devices"`
	Before  time.Time `bson:"before"`
	After   time.Time `bson:"after"`
}

// use to convey the
type SubscriptionRequest struct {
	clientId   *string
	newDevlist []string
	oldDevlist []string
}
