package main

import (
	"database/sql"
	"fmt"
	"github.com/samber/lo"
	"os"
	"testing"
)

func createTestDatabase() dbConn {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %s.\n", err.Error())
		panic(err)
	}

	err = dbInit(db, testConf.AssetsPath()+"/sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s.\n", err.Error())
		panic(err)
	}

	return db
}

func TestDbGetNetworks(t *testing.T) {
	// Arrange
	db := createTestDatabase()
	var err error
	err = dbAddServer(db, "a.foo.example.com", "foonet")
	if err != nil {
		panic(err)
	}
	err = dbAddServer(db, "b.foo.example.com", "foonet")
	if err != nil {
		panic(err)
	}
	err = dbAddServer(db, "a.bar.example.com", "barnet")
	if err != nil {
		panic(err)
	}
	err = dbAddServer(db, "b.bar.example.com", "barnet")
	if err != nil {
		panic(err)
	}

	// Act
	networks, err := dbGetNetworks(db)
	if err != nil {
		panic(err)
	}

	// Assert
	if len(networks) != 2 {
		t.Errorf("Expected two networks, but there was %d.\n", len(networks))
	}
	foonet, ok := lo.Find(networks, func(network *dbNetwork) bool { return network.network == "foonet" })
	if !ok {
		t.Fatalf("expected but could not find the foonet network.\n")
	}
	if len(foonet.servers) != 2 {
		t.Fatalf("expected foonet to have two servers, but it had %d.\n", len(foonet.servers))
	}
}
