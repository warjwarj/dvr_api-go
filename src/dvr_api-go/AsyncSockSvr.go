/*

TODO:
 - channels for accepting and releasing connections, pass channel into connHandler goroutine and trust them to
   release properly

*/

package main

import (
	"errors"
	"fmt"
	"net"
)

type AsyncSockSvr struct {
	endpoint    string         // IP + port, 123.123.123.123:1234
	capacity    int            // num of connections
	connHandler func(net.Conn) // set in the class we will embed this one into
}

// check if the struct has been instantiated properly
func (s *AsyncSockSvr) FieldsAreValid() bool {
	if s.connHandler == nil || s.endpoint == "" || s.capacity == 0 {
		return false
	}
	return true
}

func (s *AsyncSockSvr) Run() error {
	if !s.FieldsAreValid() {
		return errors.New("AsyncSockSvr wasn't initialised properly")
	}
	ln, err := net.Listen("tcp", GPS_SVR_ENDPOINT)
	if err != nil {
		return err
	} else {
		fmt.Println("Server listening on ", GPS_SVR_ENDPOINT, "...")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
		}
		go s.connHandler(conn)
	}
}
