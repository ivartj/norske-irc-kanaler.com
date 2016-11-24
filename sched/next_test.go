package sched

import (
	"testing"
	"time"
)

func TestNext(T *testing.T) {

	t, err := time.Parse(time.ANSIC, "Mon Jan 2 15:04:05 2006")
	if err != nil {
		panic(err)
	}
	t, err = NextFollowing("Sunday 15:00", t)
	if err != nil {
		panic(err)
	}
	expected := "Sun Jan  8 15:00:00 2006"
	actual := t.Format(time.ANSIC)
	if actual != expected {
		T.Errorf("'%s' did not match expected '%s'.\n", actual, expected)
	}

	t, err = time.Parse(time.ANSIC, "Mon Jan 2 15:04:05 2006")
	if err != nil {
		panic(err)
	}
	t, err = NextFollowing("Monday 15:00", t)
	if err != nil {
		panic(err)
	}
	expected = "Mon Jan  9 15:00:00 2006"
	actual = t.Format(time.ANSIC)
	if actual != expected {
		T.Errorf("'%s' did not match expected '%s'.\n", actual, expected)
	}
		
	t, err = time.Parse(time.ANSIC, "Mon Jan 2 15:04:05 2006")
	if err != nil {
		panic(err)
	}
	t, err = NextFollowing("15:00", t)
	if err != nil {
		panic(err)
	}
	expected = "Tue Jan  3 15:00:00 2006"
	actual = t.Format(time.ANSIC)
	if actual != expected {
		T.Errorf("'%s' did not match expected '%s'.\n", actual, expected)
	}
}

