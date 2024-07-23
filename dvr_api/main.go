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
	"fmt"
)

// globals
const (
	GPS_SVR_ENDPOINT string = "192.168.1.77:9047"
	DB_URL           string = "172.20.0.2:5432"
)

func main() {
	// create server
	devSvr, err := NewDeviceSvr(GPS_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println(err)
	}
	apiSvr, err := NewAPIClientSvr(GPS_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println(err)
	}
	// run servers
	go devSvr.Run()
	go apiSvr.Run()

	ConnectDB(DB_URL)

	for {
	}
	// don't like the channels for receiving data
	// each connection should
}
