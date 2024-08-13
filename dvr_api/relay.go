package main

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"nhooyr.io/websocket"
)

// record and index connected devices and clients
type Relay struct {
	logger  *zap.SugaredLogger
	devices *DeviceSvr    // dev svr
	clients *WebSockSvr   // api svr
	dbc     *DBConnection // mongodb database connection
	lock    sync.Mutex    // might be uneccessary
}

// constructor
func NewRelay(logger *zap.SugaredLogger, devices *DeviceSvr, clients *WebSockSvr, dbc *DBConnection) (*Relay, error) {
	return &Relay{
		logger:  logger,
		devices: devices,
		clients: clients,
		dbc:     dbc,
	}, nil
}

// might be able to get rid of this
type Message struct {
	message string  // text the tcp client sent
	id      *string // index which the message sender with in the connIndex of the server
}

// take messages from servers, handle, first step
func (rl *Relay) IntakeMessages() error {
	// handle messages, main program loop
	for i := 0; ; i++ {
		select {
		// process one message received from the device server
		case msgWrap, ok := <-rl.devices.svrMsgBufChan:
			if ok {
				rl.logger.Infof("Received |%q| from |%q|", msgWrap.message, *msgWrap.id)
				err := rl.ProcessMsg_Device(&msgWrap)
				if err != nil {
					rl.logger.Errorf("error processing message from device: %v", err)
				}
			} else {
				rl.logger.Error("Couldn't receive value from devMsgChan")
			}
		// process one message received from the API server
		case msgWrap, ok := <-rl.clients.svrMsgBufChan:
			if ok {
				rl.logger.Infof("Received |%q| from |%q|", msgWrap.message, *msgWrap.id)
				err := rl.ProcessMsg_APIClient(&msgWrap)
				if err != nil {
					rl.logger.Errorf("error processing message from api client: %v", err)
				}
			} else {
				rl.logger.Error("Couldn't receive value from apiMsgChan")
			}
		}
	}
}

// handle one message from an api client
func (rl *Relay) ProcessMsg_APIClient(msgWrap *Message) error {

	// verify api client connection, get the connection object
	cliConn, cliConnOk := rl.clients.connIndex.Get(*msgWrap.id)
	if !cliConnOk {
		// this really shouldn't happen because the client is almost certainly 'connected', in the connection loop, yet not registered in the controller
		return fmt.Errorf("api client not registered in controller")
	}

	// extract the device the message pertains to
	var dev_id string
	err := getIdFromMessage(&msgWrap.message, &dev_id)
	if err != nil {
		// respond to api client saying the dev id was not formatted correctly
		return fmt.Errorf("couldn't parse device id from %v", msgWrap.message)
	}

	// verify device connection, get the connection object
	devConn, devConnOk := rl.devices.connIndex.Get(dev_id)
	if !devConnOk {
		cliConn.Write(context.Background(), websocket.MessageText, []byte(fmt.Sprintf("device %v connection not registered", dev_id)))
		return nil
	}

	// send the requested message to the device
	bytes := []byte(msgWrap.message)
	sent, err := (*devConn).Write(bytes)
	if err != nil {
		return fmt.Errorf("error writing to device connection: %v", err)
	}

	// inform api client of outcome of writing to the device connection
	err = cliConn.Write(context.Background(), websocket.MessageText, []byte(fmt.Sprintf("Sent %v/%v bytes to device %v", sent, len(bytes), dev_id)))
	if err != nil {
		return fmt.Errorf("error writing to websocket connection: %v", err)
	}

	// no err
	return nil
}

// record the message in the database
func (rl *Relay) ProcessMsg_Device(msgWrap *Message) error {

	// call the insert function
	// updateResult, err := rl.dbc.RecordMessage_FromDevice(msgWrap)
	_, err := rl.dbc.RecordMessage_FromDevice(msgWrap)
	if err != nil {
		return fmt.Errorf("error recording message in db: %v", err)
	}

	// multi-faceted result, print for debug
	//rl.logger.Debug(updateResult) // MatchedCount, ModifiedCount, UpsertedCount

	// no err
	return nil
}
