/*

TODO:
	- db: postgres
	- main table for all messages. New one every week, month etc.
	- column for time received on server, raw message, and any other data we want to de-normalise from the raw message packet
	- views for each device. (Create ahead of time surely?)
	- implement view of the data on the frontend.

*/

package main

import (
	"context"
	"fmt"
)

// globals
const (
	GPS_SVR_ENDPOINT string = "192.168.1.77:9047"
	DB_URL           string = "postgres://admin:admin@172.20.0.2:5432/message_store"
)

func main() {

	// create server structs
	devSvr, err := NewDeviceSvr(GPS_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println(err)
	}
	apiSvr, err := NewAPIClientSvr(GPS_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println(err)
	}

	// start listening
	go devSvr.Run()
	go apiSvr.Run()

	// create and lightly test DB connection
	dbConn, err := NewDBConn(DB_URL, context.Background())
	err = dbConn.conn.Ping(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	// test db structure
	err = dbConn.ValidateDBStructure()
	if err != nil {
		fmt.Println(err)
		return
	}

	// store in db (hangs)
	err = dbConn.PipeMessagesToDB(devSvr.svrMsgBufChan, apiSvr.svrMsgBufChan)
	if err != nil {
		fmt.Println(err)
		return
	}
}
