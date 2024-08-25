package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func SimulateDevices(
	quantity int,
	msg_interlude int,
	errChan chan<- error,
	deviceConnections chan<- bool,
	deviceDisconnections chan<- bool,
) error {

	go func() {
		conn, err := net.Dial("tcp", GPS_SVR)
		deviceConnections <- true
		if err != nil {
			errChan <- err
			deviceDisconnections <- true
			continue
		}
		defer conn.Close()
	}()
}

func mockDevice(
	ctx context.Context,
	devId int,
	errChan chan<- error,
	deviceConnections chan<- bool,
	deviceDisconnections chan<- bool,
) {
	for {
		// wait in case of reconnect
		time.Sleep(time.Millisecond * time.Duration(CONN_INTERLUDE_MS))
		conn, err := net.Dial("tcp", GPS_SVR)
		deviceConnections <- true
		if err != nil {
			errChan <- err
			deviceDisconnections <- true
			continue
		}
		defer conn.Close()

		message := fmt.Sprintf("$ALV;%d;Hello Server!\r", devId)
		for {
			_, err = conn.Write([]byte(message))
		}
		if err != nil {
			errChan <- err
			deviceDisconnections <- true
			return
		}
		buf := make([]byte, 256)
		for {
			// just echo recvd back
			_, err := conn.Read(buf)
			if err != nil {
				errChan <- err
				deviceDisconnections <- true
				return
			}
			_, err = conn.Write(buf)
			if err != nil {
				errChan <- err
				deviceDisconnections <- true
				return
			}
		}
	}
}
