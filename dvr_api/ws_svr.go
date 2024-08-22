package main

import (
	"context"
	"net"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WebSockSvr struct {
	logger           *zap.Logger
	endpoint         string         // IP + port, ex: "192.168.1.77:9047"
	capacity         int            // num of connections
	sockOpBufSize    int            // how much memory do we give each connection to perform send/recv operations
	sockOpBufStack   Stack[*[]byte] // memory region we give each conn to so send/recv
	svrMsgBufSize    int            // how many messages can we queue on the server at once
	svrMsgBufChan    chan Message   // the channel we use to queue the messages
	svrSubReqBufChan chan SubscriptionRequest
	connIndex        Dictionary[websocket.Conn] // index the connection objects against the ids of the clients represented thusly
}

func NewWebSockSvr(logger *zap.Logger, endpoint string, capacity int, bufSize int, svrMsgBufSize int) (*WebSockSvr, error) {
	// create the struct
	svr := WebSockSvr{
		logger,
		endpoint,
		capacity,
		bufSize,
		Stack[*[]byte]{},
		svrMsgBufSize,
		make(chan Message),
		make(chan SubscriptionRequest),
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
func (s *WebSockSvr) Init() error {
	for i := 0; i < s.capacity; i++ {
		buf := make([]byte, s.sockOpBufSize)
		s.sockOpBufStack.Push(&buf)
	}
	return nil
}

// run the server
func (s *WebSockSvr) Run() {
	// listen tcp
	l, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		s.logger.Fatal("error listening on %v: %v", zap.String("s.endpoint", s.endpoint), zap.Error(err))
	} else {
		s.logger.Info("websocket server listening on: %v", zap.String("s.endpoint", s.endpoint))
	}

	// accept http on the port open for tcp above
	httpSvr := &http.Server{
		Handler: s,
	}
	err = httpSvr.Serve(l)
	if err != nil {
		s.logger.Fatal("error serving websocket server: %v", zap.Error(err))
	}
}

// func called for each connection to handle the websocket connection request, calls and blocks on connHandler
func (s *WebSockSvr) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// accept wenbsocket connection
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:       []string{"dvr_api"},
		OriginPatterns:     []string{"*"}, // Accept all origins for simplicity; customize as needed
		InsecureSkipVerify: true,          // Not recommended for production, remove this line in a real application
	})

	// handle err
	if err != nil {
		s.logger.Info("error accepting websocket connection: %v", zap.Error(err))
		return
	}
	s.logger.Info("connection accepted on api svr...")

	// don't really need this but why not
	if c.Subprotocol() != "dvr_api" {
		s.logger.Debug("declined connection because subprotocol != dvr_api")
		c.Close(websocket.StatusPolicyViolation, "client must speak the dvr_api subprotocol")
		return
	}

	// handle connection
	err = s.connHandler(c)
	if err != nil {
		s.logger.Error("error in connection handler func: %v", zap.Error(err))
	}
	c.CloseNow()
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

// handle one connection.
func (s *WebSockSvr) connHandler(conn *websocket.Conn) error {

	// msg = reusable holder for string, gen id as an arbitrary number
	var msg JsonWSMessage
	var id string = uuid.New().String()
	var subscriptions []string

	// add to connection index, defer the removal from the connection index
	s.connIndex.Add(id, *conn)
	defer s.connIndex.Delete(id)

	// connection loop
	for {
		// read one websocket message frame as json, unmarshal into a struct
		err := wsjson.Read(context.Background(), conn, &msg)
		if err != nil {
			// don't realistically need to know why but might be useful for debug.
			s.logger.Debug("websocket connection closed, status: %v", zap.String("websocket.CloseStatus(err)", websocket.CloseStatus(err).String()))
			return nil
		}

		// register the subscription request
		s.svrSubReqBufChan <- SubscriptionRequest{clientId: &id, newDevlist: msg.Subscriptions, oldDevlist: subscriptions}
		subscriptions = make([]string, len(msg.Subscriptions))
		copy(subscriptions, msg.Subscriptions)

		// todo pass the array instead of the induvidual message
		for _, val := range msg.Messages {
			s.svrMsgBufChan <- Message{val, &id}
		}
	}
}

// // handle each connection
// func (s *WebSockSvr) connHandler(conn *websocket.Conn) error {

// 	// msg = reusable holder for string, gen id as an arbitrary number
// 	var msg string
// 	var id string = uuid.New().String()

// 	// add to connection index, defer the removal from the connection index
// 	s.connIndex.Add(id, *conn)
// 	defer s.connIndex.Delete(id)

// 	// connection loop
// 	for {
// 		// read one websocket message frame
// 		wsMsgType, buf, err := conn.Read(context.Background())

// 		// if the above read operation errors then conn is closed.
// 		if err != nil {
// 			// don't realistically need to know why but might be useful for debug.
// 			s.logger.Debugf("websocket connection closed, status: %v", websocket.CloseStatus(err).String())
// 			return nil
// 		}

// 		// message should only be text
// 		if wsMsgType != websocket.MessageText {
// 			return fmt.Errorf("received non-text websocket message - closing conn")
// 		}

// 		// build message with the new data sent
// 		msg += string(buf)

// 		// check if it is a whole message
// 		if buf[len(buf)-1] != '\r' && buf[len(buf)-1] != '\n' {
// 			continue
// 		}

// 		// send message to relay and reset for next message
// 		s.svrMsgBufChan <- Message{msg, &id}
// 		msg = ""
// 	}
// 	return nil
// }