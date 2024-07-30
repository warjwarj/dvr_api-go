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
	DEVICE_SVR_ENDPOINT string = "192.168.1.77:9047"
	API_SVR_ENDPOINT    string = "192.168.1.77:9046"
	DB_URL              string = "postgresql://admin:admin@172.19.0.3:5432/message_store"
)

func main() {

	// create server structs
	devSvr, err := NewDeviceSvr(DEVICE_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println(err)
	}
	apiSvr, err := NewAPIClientSvr(API_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println(err)
	}

	// start listening
	go devSvr.Run()
	go apiSvr.Run()

	// handle messages
	for i := 0; ; i++ {
		select {
		case msgWrap, ok := <-devSvr.svrMsgBufChan:
			if ok {
				fmt.Printf("Received,|%q| from |%q|\n", *msgWrap.message, *msgWrap.id)
			} else {
				fmt.Errorf("Couldn't receive value from devMsgChan")
			}
		case msgWrap, ok := <-apiSvr.svrMsgBufChan:
			if ok {
				fmt.Printf("Received,|%q| from |%q|\n", *msgWrap.message, *msgWrap.id)
			} else {
				fmt.Errorf("Couldn't receive value from apiMsgChan")
			}
		}
	}

	// // create DB connection
	// dbConn, err := NewDBConn(DB_URL, context.Background())
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // TEST connection
	// err = dbConn.conn.Ping(context.Background())
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// // TEST db structure
	// err = dbConn.ValidateDBStructure()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// // store messages in db (hangs)
	// err = dbConn.PipeMessagesToDB(devSvr.svrMsgBufChan, apiSvr.svrMsgBufChan)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
}
