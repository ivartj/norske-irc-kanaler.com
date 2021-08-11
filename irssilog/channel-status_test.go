package irssilog

import (
	"strings"
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

const testLog string = `--- Log opened Mon Aug 15 22:11:49 2016
22:11 -!- Irssi: #example: Total of 113 nicks [6 ops, 0 halfops, 0 voices, 107 normal]
22:25 -!- FooNick [~BarUser@example-host] has joined #example
22:39 -!- FooNick [~BarUser@example-host] has quit #example
22:40 -!- BarNick [~FooUser@example-host] has left #example [Leave message]
22:41 -!- Netsplit foohost <-> barhost quits: FooNick, BarNick
22:42 -!- Netsplit over, joins: FooNick, BarNick, GooNick
22:43 -!- Netsplit foohost <-> barhost quits: FooNick, BarNick, GooNick, (+5 more, use /NETSPLIT to show all of them)
22:44 -!- Netsplit over, joins: FooNick, BarNick, GooNick, LooNick (+3 more)
22:45 -!- FooNick was kicked from #example by BarNick [Kick message]
`

func TestRegexpes(t *testing.T) {
	lines := strings.Split(testLog, "\n")
	if !regexLogOpened.MatchString(lines[0]) {
		t.Errorf("%s did not match regexp %s", lines[0], regexLogOpened)
	}
	if !regexTotalNick.MatchString(lines[1]) {
		t.Errorf("%s did not match regexp %s", lines[1], regexTotalNick)
	}
	if !regexJoined.MatchString(lines[2]) {
		t.Errorf("%s did not match regexp %s", lines[2], regexJoined)
	}
	if !regexQuit.MatchString(lines[3]) {
		t.Errorf("%s did not match regexp %s", lines[3], regexQuit)
	}
	if !regexLeft.MatchString(lines[4]) {
		t.Errorf("%s did not match regexp %s", lines[4], regexLeft)
	}
	if !regexKick.MatchString(lines[9]) {
		t.Errorf("%s did not match regexp %s", lines[9], regexKick)
	}
}

func TestGetChannelStatusFromLog(t *testing.T) {

	r := strings.NewReader(testLog)
	channelStatus, err := GetChannelStatusFromLog(r)
	if err != nil {
		t.Fatal(err)
	}

	expected := 111
	if channelStatus.NumUsers != expected {
		t.Errorf("Expected %d, not %d users in following test log:\n%s", expected, channelStatus.NumUsers, testLog)
	}

}
