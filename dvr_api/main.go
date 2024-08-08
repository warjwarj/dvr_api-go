/*

TODO:

 - do the db

*/

package main

import "fmt"

// globs
const (
	DEVICE_SVR_ENDPOINT string = "192.168.1.77:9047"        // endpoint for dev svr
	API_SVR_ENDPOINT    string = "127.0.0.1:9046"           // endpoint for api websock svr
	MONGODB_ENDPOINT    string = "mongodb://0.0.0.0:27017/" // database uri
)

func main() {

	// create server structs
	devSvr, err := NewDeviceSvr(DEVICE_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println("Error creating device server: ", err)
	}
	apiSvr, err := NewAPIClientSvr(API_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println("Error creating api server: ", err)
	}

	// create DB connection
	dbc, err := NewDBConnection(MONGODB_ENDPOINT)
	if err != nil {
		fmt.Println("Error creating database connection: ", err)
	}

	// start the servers listening
	go devSvr.Run()
	go apiSvr.Run()

	relay := Relay{devices: devSvr, clients: apiSvr, dbc: dbc}
	relay.IntakeMessages()

}
