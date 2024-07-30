package main

import (
	"fmt"
	"net"
)

type APIClientSvr struct {
	endpoint       string         // IP + port, ex: "192.168.1.77:9047"
	capacity       int            // num of connections
	sockOpBufSize  int            // how much memory do we give each connection to perform send/recv operations
	sockOpBufStack Stack[*[]byte] // memory region we give each conn to so send/recv
	svrMsgBufSize  int            // how many messages can we queue on the server at once
	svrMsgBufChan  chan Message   // the channel we use to queue the messages
}

func NewAPIClientSvr(endpoint string, capacity int, bufSize int, svrMsgBufSize int) (*APIClientSvr, error) {
	svr := APIClientSvr{
		endpoint,
		capacity,
		bufSize,
		Stack[*[]byte]{},
		svrMsgBufSize,
		make(chan Message)}
	// init the stack we use to store the buffers
	svr.sockOpBufStack.Init(capacity)
	// create and store the buffers
	for i := 0; i < svr.capacity; i++ {
		buf := make([]byte, svr.sockOpBufSize)
		svr.sockOpBufStack.Push(&buf)
	}
	return &svr, nil
}

// create and store our buffers
func (s *APIClientSvr) Init() error {
	for i := 0; i < s.capacity; i++ {
		buf := make([]byte, s.sockOpBufSize)
		s.sockOpBufStack.Push(&buf)
	}
	return nil
}

// run the server
func (s *APIClientSvr) Run() error {
	ln, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		return err
	} else {
		fmt.Println("Server listening on ", s.endpoint, "...")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
		}
		go s.connHandler(conn)
	}
}

// handle each connection
func (s *APIClientSvr) connHandler(conn net.Conn) {
	fmt.Println("Connection accepted...")
	buf, err := s.sockOpBufStack.Pop()
	if err != nil {
		fmt.Println(err)
	}
	// connection loop
	var msg string
	var id string
	for {
		recvd, err := conn.Read(*buf)
		if err != nil {
			fmt.Println("Connection closed")
			return
		} else {
			// delimit messages with the carrige return
			msg = string((*buf)[:recvd])
			if (*buf)[len((*buf))-1] == '\r' {
				if id == "" {
					err = getIdFromMessage(&msg, &id)
					if err != nil {
						fmt.Println(err)
					}
				}
				s.svrMsgBufChan <- Message{&msg, &id}
			}
		}
	}
}
