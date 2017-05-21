package irssilog

import (
	"bufio"
	"fmt"
	"regexp"
	"io"
	"strconv"
	"time"
	"strings"
)

type ChannelStatus struct{
	Time time.Time
	NumUsers int
	Topic string
}

var (
	// --- Log opened Mon Aug 15 22:11:49 2016
	regexLogOpened	= regexp.MustCompile(`^--- Log opened (.+)`)

	// --- Day changed Tue Aug 16 2016
	regexDayChanged	= regexp.MustCompile(`^--- Day changed (.+)`)

	timeFormatDayChanged = "Mon Jan _2 2006"

	// 22:11 -!- Irssi: #example: Total of 113 nicks [6 ops, 0 halfops, 0 voices, 107 normal]
	regexTotalNick	= regexp.MustCompile(`^([0-9]{2}:[0-9]{2}) -!- Irssi: #.+?: Total of ([0-9]+) nicks`)

	// 22:25 -!- FooNick [~BarUser@example-host] has joined #example
	regexJoined	= regexp.MustCompile(`^([0-9]{2}:[0-9]{2}) -!- .+? \[~.+?@.+?\] has joined`)

	// 22:39 -!- FooNick [~BarUser@example-host] has quit [Ping timeout: 246 seconds]
	regexQuit	= regexp.MustCompile(`^([0-9]{2}:[0-9]{2}) -!- .+? \[~.+?@.+?\] has quit`)

	// 18:16 -!- FooNick [~BarUser@example-host] has left #example [Leave message]
	regexLeft	= regexp.MustCompile(`^([0-9]{2}:[0-9]{2}) -!- .+? \[~.+?@.+?\] has left`)

	// 09:34 -!- FooNick changed the topic of #example to: Lorem ipsum dolor sit amet
	regexTopic	= regexp.MustCompile(`^([0-9]{2}:[0-9]{2}) -!- .+? changed the topic of .+? to: (.+)`)

	// 09:34 < FooNick> hello
	regexTimestamp	= regexp.MustCompile(`^([0-9]{2}:[0-9]{2})`)

	// 15:46 -!- Netsplit foohost <-> barhost quits: FooNick, BarNick
	regexNetsplitQuits	= regexp.MustCompile(`^([0-9]{2}:[0-9]{2}) -!- Netsplit .+? quits: (.+)`)

	// 16:00 -!- Netsplit over, joins: FooNick, BarNick (+50 more)
	regexNetsplitJoins	= regexp.MustCompile(`^([0-9]{2}:[0-9]{2}) -!- Netsplit .+? joins: (.+)`)
)


func GetChannelStatusFromLog(log io.Reader) (status ChannelStatus, err error) {

	scan := bufio.NewScanner(log)
	numusers := 0
	topic := ""
	t := time.Time{}
	date := time.Time{}

	defer func() {
		x := recover()
		if x != nil {
			return
		}
		xerr, ok := x.(error)
		if !ok {
			panic(x)
		}
		err = xerr
	}()

	for scan.Scan() {

		line := scan.Text()

		switch {

		case regexLogOpened.MatchString(line):

			submatches := xExpectSubmatches(regexLogOpened, line, 2)

			// TODO: Consider that the log is from a system with different timezone
			date, err = time.ParseInLocation(time.ANSIC, submatches[1], time.Local)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse the time given in '%s': %s", line, err.Error())
			}

			t = date

		case regexDayChanged.MatchString(line):

			submatches := xExpectSubmatches(regexDayChanged, line, 2)

			// TODO: Consider that the log is from a system with different timezone
			date, err = time.ParseInLocation(timeFormatDayChanged, submatches[1], time.Local)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse the time given in '%s': %s", line, err.Error())
			}

			t = date

		case regexTotalNick.MatchString(line):

			submatches := xExpectSubmatches(regexTotalNick, line, 3)
			strclock := submatches[1]
			strnumusers := submatches[2]

			t = xSetClock(date, strclock)

			numusers, err = strconv.Atoi(strnumusers)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse '%s' as number of total users", strnumusers)
			}

		case regexTopic.MatchString(line):
			submatches := xExpectSubmatches(regexTopic, line, 3)
			strclock := submatches[1]
			topic = submatches[2]

			t = xSetClock(date, strclock)

		case regexJoined.MatchString(line):
			submatches := xExpectSubmatches(regexJoined, line, 2)
			strclock := submatches[1]
			t = xSetClock(date, strclock)

			numusers++

		case regexQuit.MatchString(line):
			submatches := xExpectSubmatches(regexQuit, line, 2)
			strclock := submatches[1]
			t = xSetClock(date, strclock)

			numusers--

		case regexLeft.MatchString(line):
			submatches := xExpectSubmatches(regexLeft, line, 2)
			strclock := submatches[1]
			t = xSetClock(date, strclock)

			numusers--

		case regexNetsplitQuits.MatchString(line):
			submatches := xExpectSubmatches(regexNetsplitQuits, line, 3)
			strclock := submatches[1]
			quitsstr := submatches[2]

			numusers -= countNetsplitQuits(quitsstr)

			t = xSetClock(date, strclock)

		case regexNetsplitJoins.MatchString(line):
			submatches := xExpectSubmatches(regexNetsplitJoins, line, 3)
			strclock := submatches[1]
			joinsstr := submatches[2]

			numusers += countNetsplitJoins(joinsstr)

			t = xSetClock(date, strclock)

		// Must be last case
		case regexTimestamp.MatchString(line):
			submatches := xExpectSubmatches(regexTimestamp, line, 2)
			strclock := submatches[1]
			t = xSetClock(date, strclock)

		}

	}

	if scan.Err() != nil {
		return ChannelStatus{}, fmt.Errorf("Error on scanning line: %s", scan.Err().Error())
	}

	if t.IsZero() {
		return ChannelStatus{}, fmt.Errorf("No specific time was found throughout the log")
	}

	return ChannelStatus{Time: t, NumUsers: numusers, Topic: topic}, nil
}

func xErrorf(format string, args... interface{}) {
	panic(fmt.Errorf(format, args...))
}

func xExpectSubmatches(regex *regexp.Regexp, line string, nsubmatches int) (submatches []string) {
	submatches = regex.FindStringSubmatch(line)
	if len(submatches) == nsubmatches {
		xErrorf("Expected %d submatches for line '%s' by regexp '%s'", nsubmatches, line, regex.String())
	}
	return submatches
}

// Given a string that may or may not contain "(+<num> more", returns <num> or
// 0 if no such ending is found
func howManyMore(str string) int {
	submatches := regexHowManyMore.FindStringSubmatch(str)
	if len(submatches) != 2 {
		return 0
	}
	num, err := strconv.Atoi(submatches[1])
	if err != nil {
		return 0
	}
	return num
}
var regexHowManyMore = regexp.MustCompile(`.+? \(\+([0-9]+) more`)

func xSetClock(date time.Time, strclock string) time.Time {
	var hour, min int
	_, err := fmt.Sscanf(strclock, "%d:%d", &hour, &min)
	if err != nil {
		xErrorf("Failed to scan '%s' as clock string: %s", strclock, err.Error())
	}
	t := date.Add(time.Hour * time.Duration(hour) + time.Minute * time.Duration(min))
	return t
}

func countNetsplitQuits(quitstr string) int {
	// FooNick, BarNick
	// FooNick, BarNick, (+5 more, use /NETSPLIT to show all of them)

	ncommas := strings.Count(quitstr, ",")

	// Check for 'more' parentheses
	more := howManyMore(quitstr)
	if more != 0 {
		// Parentheses will contain a comma, hence - 1
		return ncommas + more - 1
	} else {
		return ncommas + 1
	}
}

func countNetsplitJoins(joinsstr string) int {
	// FooNick, BarNick
	// FooNick, BarNick (+50 more)

	ncommas := strings.Count(joinsstr, ",")
	more := howManyMore(joinsstr)
	if more != 0 {
		return ncommas+1 + more
	} else {
		return ncommas+1
	}
}

