package irssilog

import (
	"testing"
)

func TestCountNetsplitQuits(t *testing.T) {
	
	str1 := "FooNick, BarNick"
	str2 := "FooNick, BarNick, (+5 more, use /NETSPLIT to show all of them)"

	count1 := countNetsplitQuits(str1)
	if count1 != 2 {
		t.Errorf("'%s' was counted as %d", str1, count1)
	}

	count2 := countNetsplitQuits(str2)
	if count2 != 7 {
		t.Errorf("'%s' was counted as %d", str2, count2)
	}

}

func TestCountNetsplitJoins(t *testing.T) {

	str1 := "FooNick, BarNick"
	str2 := "FooNick, BarNick (+50 more)"

	count1 := countNetsplitJoins(str1)
	if count1 != 2 {
		t.Errorf("'%s' was counted as %d", str1, count1)
	}

	count2 := countNetsplitJoins(str2)
	if count2 != 52 {
		t.Errorf("'%s' was counted as %d", str2, count2)
	}
}

