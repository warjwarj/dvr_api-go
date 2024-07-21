/*

TODO:
 - add cache
 - make the API server compatible with websockets
 - internal dict to each server that you can use to match client identifiers to their sockets and send messages - channel to pipe the messages into the server then handle the send internally
*/

package main

import (
	"fmt"
)

// globals
var (
	GPS_SVR_ENDPOINT string = "192.168.1.77:9047"
)

func main() {
	// create server
	devSvr, err := NewDeviceSvr(GPS_SVR_ENDPOINT, 5, 1024, 40)
	if err != nil {
		fmt.Println(err)
	}
	// run server
	go devSvr.Run()
	// handle data received from devices connected to the server
	for {
		select {
		case msg, ok := <-devSvr.svrMsgBufChan:
			if ok {
				// send this to a DB or cache or something
				fmt.Println(msg)
			} else {
				fmt.Println("message buffer chan shouldn't be closed")
			}
		}
	}
}
