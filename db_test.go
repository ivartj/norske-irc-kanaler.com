package main

import (
	"database/sql"
	"fmt"
	"github.com/ivartj/norske-irc-kanaler.com/util"
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
	foonets := util.Filter(networks, func(network *dbNetwork) bool { return network.network == "foonet" })
	if len(foonets) == 0 {
		t.Fatalf("expected but could not find the foonet network.\n")
	}
	foonet := foonets[0]
	if len(foonet.servers) != 3 {
		t.Errorf("expected foonet to have three servers, but it had %d.\n", len(foonet.servers))
		t.Logf("foonet servers:\n")
		for _, server := range foonet.servers {
			t.Logf(" - %s\n", server)
		}
	}
}
