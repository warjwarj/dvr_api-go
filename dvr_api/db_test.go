package main

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// test the new db connection
func TestDatabaseConnection(t *testing.T) {
	dbc, err := NewDBConnection(MONGODB_ENDPOINT)
	if err != nil {
		t.Fatalf("NewDBConnection(uri) directly returned error: %v", err)
	}
	if dbc == nil {
		t.Fatalf("db client obj returned from NewDBConnection(uri) was nil")
	}
	// Insert a test document into the database
	collection := dbc.client.Database("testdb").Collection("people")
	// ins
	insRes, err := collection.InsertOne(context.Background(), bson.M{"Name": "Alexander Hypocroties Bivouthigronaties"})
	if err != nil {
		t.Fatalf("error inserting into test db: %v", err)
	}
	if insRes == nil {
		t.Fatal("insert result was nil")
	}
	// del
	delRes, err := collection.DeleteOne(context.Background(), bson.M{"Name": "Alexander Hypocroties Bivouthigronaties"})
	if err != nil {
		t.Fatalf("error deleting from test db: %v", err)
	}
	if delRes == nil {
		t.Fatal("delete result was nil")
	}
}
