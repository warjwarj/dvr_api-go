package main

import (
	"time"
)

// used in ws_svr.go to read json messages into structs
type ApiReq_WS struct {
	Messages            []string `json:"messages"`
	Subscriptions       []string `json:"subscriptions"`
	GetConnectedDevices bool     `json:"getConnectedDevices"`
}

// used in ws_svr.go to send a websocket message containing all
type ApiRes_ConnectedDevicesList_WS struct {
	ConnectedDevicesList []string `json:"connectedDevicesList"`
}

// pass messages out of server into handlers
type Message struct {
	message   string    // text the tcp client sent
	clientId  *string   // index which the message sender with in the connIndex of the server
	recvdTime time.Time // recvd time
}

// Device schema for modelling in mongodb
type Device_Schema struct {
	DeviceId   string                 `bson:"device_id"`
	MsgHistory []DeviceMessage_Schema `bson:"msg_history"`
}

// Device message schema for modelling in mongodb
// Also used for responding to API clients
type DeviceMessage_Schema struct {
	RecvdTime  time.Time `bson:"received_time"`
	PacketTime time.Time `bson:"packet_time"`
	Message    string    `bson:"message"`
	Direction  string    `bson:"direction"`
}

// time is in the ISO string format
type ApiRequest_HTTP struct {
	Devices []string  `bson:"devices"`
	Before  time.Time `bson:"before"`
	After   time.Time `bson:"after"`
}

// use to convey subscription requests to the handler from the server
type SubscriptionRequest struct {
	clientId   *string
	newDevlist []string
	oldDevlist []string
}
