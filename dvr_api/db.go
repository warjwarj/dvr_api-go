package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

type DBConn struct {
	conn *pgx.Conn
}

// construct
func NewDBConn(url string, ctx context.Context) (*DBConn, error) {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	return &DBConn{conn: conn}, nil

}

// This function validates tables, views etc. Tries to fix, errors if can't
func (db *DBConn) ValidateDBStructure() error {
	fmt.Println("implement this")
	return nil
}

// intake messages
func (db *DBConn) PipeMessagesToDB(devMsgChan <-chan string, apiMsgChan <-chan string) error {
	for i := 0; ; i++ {
		select {
		case msg, ok := <-devMsgChan:
			if ok {
				fmt.Println("Received, storing message: ", msg)

			} else {
				fmt.Errorf("Couldn't receive value from devMsgChan")
			}
		case msg, ok := <-apiMsgChan:
			if ok {
				fmt.Println("Received, storing message: ", msg)
			} else {
				fmt.Errorf("Couldn't receive value from apiMsgChan")
			}
		}
	}
}
