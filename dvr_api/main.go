package main

import (
	"fmt"
	"net"
)

// globals
var (
	GPS_SVR_ENDPOINT string = "192.168.1.77:9047"
)

func main() {
	svr := &AsyncSockSvr{
		endpoint:    GPS_SVR_ENDPOINT,
		capacity:    3,
		connHandler: handleConnection,
	}
	err := svr.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("Connection accepted...")
	buf := make([]byte, 256)
	for {
		recvd, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		} else {
			msg := string(buf[:recvd])
			fmt.Println(msg)
		}
	}
}
