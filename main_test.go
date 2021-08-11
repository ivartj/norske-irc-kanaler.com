package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	cfg := confNew()
	err := cfg.ParseFile("config.cfg.example")
	cfg.set.DatabasePath = ":memory:"
	if err != nil {
		t.Fatalf("Error on parsing example file: %s.\n", err.Error())
	}

	mainNewContext(cfg)
}
