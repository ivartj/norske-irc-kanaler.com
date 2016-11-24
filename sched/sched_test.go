package sched

import (
	"testing"
	"time"
	"fmt"
)

func testFn() {
	fmt.Println("Hello world!")
}

func TestScheduler(t *testing.T) {

	s := New()
	s.Schedule(testFn, time.Now(), time.Second * time.Duration(2))
	go s.Run()
	time.Sleep(time.Second * time.Duration(5))
	s.Stop()

}

