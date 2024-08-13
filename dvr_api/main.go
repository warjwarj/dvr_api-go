/*

TODO:

 - do the db

*/

package main

import (
	"go.uber.org/zap"
)

// Globals
const (
	// endopints
	DEVICE_SVR_ENDPOINT  string = "192.168.1.77:9047"        // endpoint for dev svr
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
	var logger *zap.SugaredLogger
	if PROD {
		tmp, _ := zap.NewProduction()
		logger = tmp.Sugar()
	} else {
		tmp, _ := zap.NewDevelopment()
		logger = tmp.Sugar()
	}
	defer logger.Sync() // flushes buffer, if any

	// create DB connection
	dbc, err := NewDBConnection(logger, MONGODB_ENDPOINT)
	if err != nil {
		logger.Fatalf("fatal error creating database connection: %v", err)
	}

	// create device server struct
	devSvr, err := NewDeviceSvr(logger, DEVICE_SVR_ENDPOINT, CAPACITY, BUF_SIZE, SVR_MSGBUF_SIZE)
	if err != nil {
		logger.Fatalf("fatal error creating device server: %v", err)
	}

	// create ws server struct
	wsSvr, err := NewWebSockSvr(logger, WEBSOCK_SVR_ENDPOINT, CAPACITY, BUF_SIZE, SVR_MSGBUF_SIZE)
	if err != nil {
		logger.Fatalf("fatal error creating api server: %v", err)
	}

	// create http server struct
	httpSvr, err := NewRestSvr(logger, REST_SVR_ENDPOINT, dbc)
	if err != nil {
		logger.Fatalf("fatal error creating REST api server: %v", err)
	}

	// start the servers listening
	go devSvr.Run()
	go wsSvr.Run()
	go httpSvr.Run()

	// create the 'relay' struct, start the intake of the messages
	relay, err := NewRelay(logger, devSvr, wsSvr, dbc)
	if err != nil {
		logger.Fatalf("fatal error creating relay struct: %v", err)
	}

	// this is the main program loop
	err = relay.IntakeMessages()
	if err != nil {
		logger.Fatalf("fatal error in IntakeMessages(): %v", err)
	}
}
