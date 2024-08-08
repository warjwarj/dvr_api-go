package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type APIClientSvr struct {
	endpoint       string                     // IP + port, ex: "192.168.1.77:9047"
	capacity       int                        // num of connections
	sockOpBufSize  int                        // how much memory do we give each connection to perform send/recv operations
	sockOpBufStack Stack[*[]byte]             // memory region we give each conn to so send/recv
	svrMsgBufSize  int                        // how many messages can we queue on the server at once
	svrMsgBufChan  chan Message               // the channel we use to queue the messages
	connIndex      Dictionary[websocket.Conn] // index the connection objects against the ids of the clients represented thusly
}

func NewAPIClientSvr(endpoint string, capacity int, bufSize int, svrMsgBufSize int) (*APIClientSvr, error) {
	// create the struct
	svr := APIClientSvr{
		endpoint,
		capacity,
		bufSize,
		Stack[*[]byte]{},
		svrMsgBufSize,
		make(chan Message),
		Dictionary[websocket.Conn]{}}
	// init things that need initing
	svr.sockOpBufStack.Init()
	svr.connIndex.Init()
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
func (acs *APIClientSvr) Run() {
	l, err := net.Listen("tcp", acs.endpoint)
	if err != nil {
		fmt.Println(err)
	}
	httpSvr := &http.Server{
		Handler: acs,
	}
	errc := make(chan error, 1)
	go func() {
		errc <- httpSvr.Serve(l)
	}()
	fmt.Printf("Websocket server listening on: %v\n", acs.endpoint)
}

func (s *APIClientSvr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:       []string{"dvr_api"},
		OriginPatterns:     []string{"*"}, // Accept all origins for simplicity; customize as needed
		InsecureSkipVerify: true,          // Not recommended for production, remove this line in a real application
	})
	fmt.Println("Connection accepted...")
	if err != nil {
		fmt.Println(err)
		return
	}
	if c.Subprotocol() != "dvr_api" {
		c.Close(websocket.StatusPolicyViolation, "client must speak the dvr_api subprotocol")
		return
	}
	go func() {
		err := s.connHandler(c)
		if err != nil {
			fmt.Println("connection handler error:", err)
		}
		c.CloseNow()
	}()
}

/*
 *  %x0 denotes a continuation frame
 *  %x1 denotes a text frame
 *  %x2 denotes a binary frame
 *  %x3-7 are reserved for further non-control frames
 *  %x8 denotes a connection close
 *  %x9 denotes a ping
 *  %xA denotes a pong
 *  %xB-F are reserved for further control frames
 */

// handle each connection
func (s *APIClientSvr) connHandler(conn *websocket.Conn) error {

	// msg = reusable holder for string, gen id as an arbitrary number
	var msg string
	var id string = uuid.New().String()

	// add to connection index, defer the removal from the connection index
	s.connIndex.Add(id, *conn)
	defer s.connIndex.Delete(id)

	// connection loop
	for {
		// read one websocket message frame
		wsMsgType, buf, err := conn.Read(context.Background())

		// handle errors
		if err != nil {
			return fmt.Errorf("Connection closed: %v\n", err)
		}
		if wsMsgType != websocket.MessageText {
			return fmt.Errorf("received non-text websocket message - closing conn")
		}
		msg += string(buf)

		// check if it is a whole message
		if buf[len(buf)-1] != '\r' && buf[len(buf)-1] != '\n' {
			continue
		}

		// Queue message for processsing
		s.svrMsgBufChan <- Message{msg, &id}
		msg = ""
	}
}
