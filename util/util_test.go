package util

import (
	"fmt"
	"golang.org/x/exp/slices"
	"testing"
)

func TestMap(t *testing.T) {
	is := []int{1, 2, 3}

	ss := Map(is, func(i int) string { return fmt.Sprintf("%d", i) })

	expected := []string{"1", "2", "3"}
	if !slices.Equal(ss, expected) {
		t.Errorf("expected %v, got %v\n", expected, ss)
	}
}

func TestGroupBy(t *testing.T) {
	ss := []string{"fool", "foot", "barbara", "barber", "bazaar"}

	m := GroupBy(ss, func(s string) string { return s[:3] })

	if len(m) != 3 {
		t.Errorf("expected 3 keys, got %d\n", len(m))
	}
	for key, expected := range map[string][]string{
		"foo": []string{"fool", "foot"},
		"bar": []string{"barbara", "barber"},
		"baz": []string{"bazaar"},
	} {
		if !slices.Equal(m[key], expected) {
			t.Errorf("expected %v, got %v\n", expected, m[key])
		}
	}
}

func TestFilter(t *testing.T) {
	is := []int{1, 2, 3, 4, 5, 6}

	odd := Filter(is, func(i int) bool { return i%2 != 0 })

	expected := []int{1, 3, 5}
	if !slices.Equal(odd, expected) {
		t.Errorf("expected %v, got %v\n", expected, odd)
	}
}
