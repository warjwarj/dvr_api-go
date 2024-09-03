package main

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// this is meant for the publish function in the sub handler.
type PublishFunction func(*MessageWrapper) error

// record and index connected devices and clients
type MessageHandler struct {
	// internal
	lock sync.Mutex // might be uneccessary

	// injected
	logger  *zap.Logger
	devices *DeviceSvr      // dev svr
	clients *WebSockSvr     // api svr
	dbc     *DBConnection   // mongodb database connection
	publish PublishFunction // this func is meant to publish a message to subscribers
}

// constructor
func NewMessageHandler(logger *zap.Logger, devices *DeviceSvr, clients *WebSockSvr, dbc *DBConnection, publish PublishFunction) (*MessageHandler, error) {
	r := &MessageHandler{
		logger:  logger,
		devices: devices,
		clients: clients,
		dbc:     dbc,
		publish: publish,
	}
	return r, nil
}

// take messages from servers, handle, first step
func (mh *MessageHandler) MsgIntake() error {
	// handle messages, main program loop
	for i := 0; ; i++ {
		select {
		// process one message received from the device server
		case msgWrap, ok := <-mh.devices.svrMsgBufChan:
			if ok {
				mh.logger.Info("Processing message", zap.String("msgWrap.message", msgWrap.message), zap.String("*msgWrap.clientId", *msgWrap.clientId))
				err := mh.ProcessMsgFromDevice(&msgWrap)
				if err != nil {
					mh.logger.Error("error processing message from device", zap.Error(err))
				}
			} else {
				mh.logger.Error("Couldn't receive value from devMsgChan")
			}
		// process one message received from the API server
		case msgWrap, ok := <-mh.clients.svrMsgBufChan:
			if ok {
				mh.logger.Info("Processing message", zap.String("msgWrap.message", msgWrap.message), zap.String("*msgWrap.clientId", *msgWrap.clientId))
				err := mh.ProcessMsgFromApiClient(&msgWrap)
				if err != nil {
					mh.logger.Error("error processing message from api client", zap.Error(err))
				}
			} else {
				mh.logger.Error("Couldn't receive value from apiMsgChan")
			}
		}
	}
}

// handle one message from an api client
func (mh *MessageHandler) ProcessMsgFromApiClient(msgWrap *MessageWrapper) error {

	// extract the device the message pertains to
	var dev_id string
	err := getIdFromMessage(&msgWrap.message, &dev_id)
	if err != nil {
		return fmt.Errorf("couldn't parse device id from %v", msgWrap.message)
	}

	// verify device connection, get the connection object
	devConn, devConnOk := mh.devices.connIndex.Get(dev_id)
	if !devConnOk {
		return fmt.Errorf("message sent for device not connected: %v", err)
	}

	// send the requested message to the device
	_, err = (*devConn).Write([]byte(msgWrap.message))
	if err != nil {
		return fmt.Errorf("error writing to device connection: %v", err)
	}

	// record message in database
	_, err = mh.dbc.RecordMessage_ToFromDevice(false, msgWrap) // MatchedCount, ModifiedCount, UpsertedCount
	if err != nil {
		return fmt.Errorf("error recording message in db: %v", err)
	}

	// no err
	return nil
}

// record the message in the database
func (mh *MessageHandler) ProcessMsgFromDevice(msgWrap *MessageWrapper) error {

	// record message in database
	_, err := mh.dbc.RecordMessage_ToFromDevice(true, msgWrap) // MatchedCount, ModifiedCount, UpsertedCount
	if err != nil {
		return fmt.Errorf("error recording message in db: %v", err)
	}

	// publish the message
	err = mh.publish(msgWrap)
	if err != nil {
		return fmt.Errorf("error publishing message: %v", err)
	}

	// no err
	return nil
}
