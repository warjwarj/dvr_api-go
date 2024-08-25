package main

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"nhooyr.io/websocket/wsjson"
)

// record and index connected devices and clients
type SubscriptionHandler struct {

	// internal
	subscriptions map[string]map[string]string // device ids against connections. Use internal map just for indexing the keys
	lock          sync.Mutex                   // might be uneccessary

	// injected
	logger  *zap.Logger
	devices *DeviceSvr    // dev svr
	clients *WebSockSvr   // api svr
	dbc     *DBConnection // mongodb database connection
}

// constructor
func NewSubscriptionHandler(logger *zap.Logger, devices *DeviceSvr, clients *WebSockSvr, dbc *DBConnection) (*SubscriptionHandler, error) {
	r := &SubscriptionHandler{
		logger:        logger,
		devices:       devices,
		clients:       clients,
		dbc:           dbc,
		subscriptions: make(map[string]map[string]string),
	}
	return r, nil
}

func (sh *SubscriptionHandler) SubIntake() error {
	// handle messages, main program loop
	for i := 0; ; i++ {
		// process one message received from the API server
		subReq, ok := <-sh.clients.svrSubReqBufChan
		if ok {
			err := sh.Subscribe(&subReq)
			if err != nil {
				sh.logger.Error("error processing subscription request: %v", zap.Error(err))
			}
		} else {
			sh.logger.Error("Couldn't receive value from svrSubReqBufChan")
		}
	}
}

// add the subscription requester's connection onto the list of subscribers for each device
func (sh *SubscriptionHandler) Subscribe(subReq *SubscriptionRequest) error {
	// delete old subs
	for _, val := range subReq.oldDevlist {
		// if the map that holds subs isn't inited then just continue
		if sh.subscriptions[val] == nil {
			continue
		}
		delete(sh.subscriptions[val], *subReq.clientId)
	}
	// add new subs
	for _, val := range subReq.newDevlist {
		// if the requested device isn't currently registered as a 'publisher'
		// then register them and add this client as a subscriber in case they register, connect, in the future
		if sh.subscriptions[val] == nil {
			sh.subscriptions[val] = make(map[string]string, 0)
		}
		sh.subscriptions[val][*subReq.clientId] = *subReq.clientId
	}
	return nil
}

// publish a message. This function works
func (sh *SubscriptionHandler) Publish(msgWrap *Message) error {
	// check if there are even eny susbcribers
	if sh.subscriptions[*msgWrap.clientId] == nil {
		return nil
	}
	// broadcast message to subscribers
	for k, _ := range sh.subscriptions[*msgWrap.clientId] {
		// get the websocket connection associated with the id stored in the sub list
		conn, ok := sh.clients.connIndex.Get(k)
		if !ok {
			// if client doesn't exist remove its entry in the map and continue
			delete(sh.subscriptions[*msgWrap.clientId], k)
			continue
		}
		packTime, err := getDateFromMessage(msgWrap.message)
		devMsg := &DeviceMessage_Schema{msgWrap.recvdTime, packTime, msgWrap.message, "from"}
		// send message to this subscriber
		err = wsjson.Write(context.TODO(), conn, devMsg)
		if err != nil {
			delete(sh.subscriptions[*msgWrap.clientId], k)
			sh.logger.Debug("removed subscriber %v from subscription list because a write operation failed", zap.String("k", k))
			continue
		}
	}
	// no err
	return nil
}
