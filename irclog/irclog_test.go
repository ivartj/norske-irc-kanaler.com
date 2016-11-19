package irclog

import (
	"testing"
	"os"
)

func TestNumUsers(t *testing.T) {
	f, err := os.Open("#scene.no.log")
	if err != nil {
		panic(err)
	}
	NumUsers(f)
	f.Close()
}

