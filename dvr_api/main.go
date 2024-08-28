/*

TODO:


*/

package main

import (
	"fmt"

	"go.uber.org/zap"
)

// Globals
const (
	// endopints
	DEVICE_SVR_ENDPOINT  string = "127.0.0.1:9047"           // endpoint for dev svr
	WEBSOCK_SVR_ENDPOINT string = "127.0.0.1:9046"           // endpoint for api websock svr
	REST_SVR_ENDPOINT    string = "127.0.0.1:9045"           // endpoint for api REST svr
	MONGODB_ENDPOINT     string = "mongodb://0.0.0.0:27017/" // database uri

	// just use this for the logger atm
	PROD bool = false

	// server configuration variables
	CAPACITY        int = 20   // how many devices can connect to the server
	BUF_SIZE        int = 1024 // how much memory will you allocate to IO operations
	SVR_MSGBUF_SIZE int = 40   // capacity of message queue
)

func main() {
	// set up our logger
	var logger *zap.Logger
	if PROD {
		tmp, err := zap.NewProduction()
		if err != nil {
			fmt.Println("error initing logger: ", err)
		}
		logger = tmp
	} else {
		tmp, err := zap.NewDevelopment()
		if err != nil {
			fmt.Println("error initing logger: ", err)
		}
		logger = tmp
	}
	defer logger.Sync() // flushes buffer, if any

	// create DB connection
	dbc, err := NewDBConnection(logger, MONGODB_ENDPOINT, "dvr_api-GPS-DB")
	if err != nil {
		logger.Fatal("fatal error creating database connection: %v", zap.Error(err))
	}

	// create device server struct
	devSvr, err := NewDeviceSvr(logger, DEVICE_SVR_ENDPOINT, CAPACITY, BUF_SIZE, SVR_MSGBUF_SIZE)
	if err != nil {
		logger.Fatal("fatal error creating device server: %v", zap.Error(err))
	}

	// create ws server struct
	wsSvr, err := NewWebSockSvr(logger, WEBSOCK_SVR_ENDPOINT, CAPACITY, BUF_SIZE, SVR_MSGBUF_SIZE, devSvr.connIndex.GetAllKeys)
	if err != nil {
		logger.Fatal("fatal error creating api server: %v", zap.Error(err))
	}

	// create http server struct
	httpSvr, err := NewHttpSvr(logger, REST_SVR_ENDPOINT, dbc)
	if err != nil {
		logger.Fatal("fatal error creating REST api server: %v", zap.Error(err))
	}

	// start the servers listening
	go devSvr.Run()
	go wsSvr.Run()
	go httpSvr.Run()

	// create the 'relay' struct, start the intake of the messages
	subHandler, err := NewSubscriptionHandler(logger, devSvr, wsSvr, dbc)
	if err != nil {
		logger.Fatal("fatal error creating relay struct: %v", zap.Error(err))
	}

	// create the 'relay' struct, start the intake of the messages. Inject the publish function into the handler struct
	msgHandler, err := NewMessageHandler(logger, devSvr, wsSvr, dbc, subHandler.Publish)
	if err != nil {
		logger.Fatal("fatal error creating relay struct: %v", zap.Error(err))
	}

	// handle messages
	go func() {
		err = msgHandler.MsgIntake()
		if err != nil {
			logger.Fatal("fatal error in MessageHandler(): %v", zap.Error(err))
		}
	}()

	// handle subscriptions
	err = subHandler.SubIntake()
	if err != nil {
		logger.Fatal("fatal error in SubsciptionHandler(): %v", zap.Error(err))
	}
}
