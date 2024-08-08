package main

import (
	"context"
	"fmt"
	"sync"

	"nhooyr.io/websocket"
)

type Message struct {
	message string  // text the tcp client sent
	id      *string // index which the message sender with in the connIndex of the server
}

// record and index connected devices and clients
type Relay struct {
	devices *DeviceSvr    // dev svr
	clients *APIClientSvr // api svr
	dbc     *DBConnection // mongodb database connection
	lock    sync.Mutex    // might be uneccessary
}

// take messages from servers, handle, first step
func (rl *Relay) IntakeMessages() {
	// handle messages, main program loop
	for i := 0; ; i++ {
		select {
		// process one message received from the device server
		case msgWrap, ok := <-rl.devices.svrMsgBufChan:
			if ok {
				fmt.Printf("Received |%q| from |%q|\n", msgWrap.message, *msgWrap.id)
				err := rl.ProcessMsg_Device(&msgWrap)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Errorf("Couldn't receive value from devMsgChan")
			}
		// process one message received from the API server
		case msgWrap, ok := <-rl.clients.svrMsgBufChan:
			if ok {
				fmt.Printf("Received |%q| from |%q|\n", msgWrap.message, *msgWrap.id)
				err := rl.ProcessMsg_APIClient(&msgWrap)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Errorf("Couldn't receive value from apiMsgChan")
			}
		}
	}
}

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

// simple conpared to the above, all we want to do is record the message in the database
func (rl *Relay) ProcessMsg_Device(msgWrap *Message) error {

	// call the insert function
	updateResult, err := rl.dbc.RecordMessage_FromDevice(msgWrap)
	if err != nil {
		return fmt.Errorf("error recording message in db: %v", err)
	}

	// multi-faceted result, print for debug
	fmt.Println(updateResult) // MatchedCount, ModifiedCount, UpsertedCount

	// no err
	return nil
}
