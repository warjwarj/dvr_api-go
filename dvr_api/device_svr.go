package main

import (
	"fmt"
	"net"
)

type DeviceSvr struct {
	endpoint       string               // IP + port, ex: "192.168.1.77:9047"
	capacity       int                  // num of connections
	sockOpBufSize  int                  // how much memory do we give each connection to perform send/recv operations
	sockOpBufStack Stack[*[]byte]       // memory region we give each conn to so send/recv
	svrMsgBufSize  int                  // how many messages can we queue on the server at once
	svrMsgBufChan  chan Message         // the channel we use to queue the messages
	connIndex      Dictionary[net.Conn] // index the connection objects against the ids of the devices represented thusly
}

func NewDeviceSvr(endpoint string, capacity int, bufSize int, svrMsgBufSize int) (*DeviceSvr, error) {
	svr := DeviceSvr{
		endpoint,
		capacity,
		bufSize,
		Stack[*[]byte]{},
		svrMsgBufSize,
		make(chan Message),
		Dictionary[net.Conn]{}}
	// init the stack we use to store the buffers
	svr.sockOpBufStack.Init()
	// init the Dictionary
	svr.connIndex.Init()
	// create and store the buffers
	for i := 0; i < svr.capacity; i++ {
		buf := make([]byte, svr.sockOpBufSize)
		svr.sockOpBufStack.Push(&buf)
	}

	return &svr, nil
}

// create and store our buffers
func (s *DeviceSvr) Init() error {
	for i := 0; i < s.capacity; i++ {
		buf := make([]byte, s.sockOpBufSize)
		s.sockOpBufStack.Push(&buf)
	}
	return nil
}

// run the server
func (s *DeviceSvr) Run() {
	ln, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Server listening on ", s.endpoint, "...")
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
		}
		go func() {
			err := s.connHandler(c)
			if err != nil {
				fmt.Println(err)
			}
			c.Close()
		}()
	}
}

// handle each connection
func (s *DeviceSvr) connHandler(conn net.Conn) error {
	fmt.Println("Connection accepted...")
	buf, err := s.sockOpBufStack.Pop()
	if err != nil {
		fmt.Println(err)
	}
	// connection loop
	var msg string = ""
	var id string = ""
	var recvd int = 0
	for {
		// read from wherever we finished last time
		tmp, err := conn.Read((*buf)[recvd:])
		recvd += tmp
		if err != nil {
			fmt.Println("Connection closed")
			s.sockOpBufStack.Push(buf)
			if id != "" {
				fmt.Println(s.connIndex)
			}
			return err
		}
		// check if we've recvd complete message.
		if (*buf)[recvd-1] != '\r' {
			fmt.Println((*buf)[recvd])
			continue
		}
		// get complete message, reset byte counter
		msg = string((*buf)[:recvd])
		recvd = 0
		// set id if needed
		if id == "" {
			err = getIdFromMessage(&msg, &id)
			if err != nil {
				fmt.Println(err)
				continue
			}
			s.connIndex.Add(id, conn)
			defer s.connIndex.Delete(id)
			fmt.Println(s.connIndex)
		}
		s.svrMsgBufChan <- Message{msg, &id}
	}
}
