package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DBConn struct {
	conn *pgx.Conn
}

// construct
func NewDBConn(url string, ctx context.Context) (*DBConn, error) {
	conn, err := pgx.Connect(ctx, url)
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
