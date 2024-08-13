package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type httpSvr struct {
	logger   *zap.SugaredLogger
	endpoint string // IP + port, ex: "192.168.1.77:9047"
	dbc      *DBConnection
}

func NewRestSvr(logger *zap.SugaredLogger, endpoint string, dbc *DBConnection) (*httpSvr, error) {
	// create the struct
	svr := httpSvr{
		logger,
		endpoint,
		dbc,
	}
	return &svr, nil
}

type APIReq struct {
	Before  time.Time `bson:"before"`
	After   time.Time `bson:"after"`
	Devices []string  `bson:"devices"`
}

// run the server
func (s *httpSvr) Run() {
	// listen tcp
	l, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		s.logger.Fatalf("error listening on %v: %v", s.endpoint, err)
	} else {
		s.logger.Infof("http api server listening on: %v", s.endpoint)
	}

	// accept http on the port open for tcp above
	httpSvr := &http.Server{
		Handler: s,
	}
	err = httpSvr.Serve(l)
	if err != nil {
		s.logger.Fatalf("error serving http api server: %v", err)
	}
}

// serve the http API
func (s *httpSvr) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Reading the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Errorf("Unable to read request body", http.StatusBadRequest)
		return
	}

	var req APIReq
	json.Unmarshal(body, &req)

	fmt.Println(req.After)
	fmt.Println(req.Before)
	fmt.Println(req.Devices)

	//res, err = s.dbc.QueryMsgHistory()
}
