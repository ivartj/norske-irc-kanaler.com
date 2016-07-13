package main

import (
	"testing"
)

func TestLoginPage(t *testing.T) {
	ctx := mainNewContext(testConf)
	_, err := testLogin(ctx)
	if err != nil {
		t.Fatalf("Failed to log in: %s.\n", err.Error())
	}
}

