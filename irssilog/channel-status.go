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


func GetChannelStatusFromLog(log io.Reader) (ChannelStatus, error) {

	var err error

	scan := bufio.NewScanner(log)
	numusers := 0
	topic := ""
	t := time.Time{}
	date := time.Time{}

	for scan.Scan() {

		line := scan.Text()

		switch {

		case regexLogOpened.MatchString(line):

			submatches := regexLogOpened.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Failed to capture by regex the date from the line '%s'", line)
			}

			// TODO: Consider that the log is from a system with different timezone
			date, err = time.ParseInLocation(time.ANSIC, submatches[1], time.Local)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse the time given in '%s': %s", line, err.Error())
			}

			t = date

		case regexDayChanged.MatchString(line):

			submatches := regexDayChanged.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Failed to capture by regex the date given in the line '%s'", line)
			}

			// TODO: Consider that the log is from a system with different timezone
			date, err = time.ParseInLocation(timeFormatDayChanged, submatches[1], time.Local)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse the time given in '%s': %s", line, err.Error())
			}

			t = date

		case regexTotalNick.MatchString(line):

			submatches := regexTotalNick.FindStringSubmatch(line)
			if len(submatches) != 3 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			strnumusers := submatches[2]

			t, err = setClock(date, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

			numusers, err = strconv.Atoi(strnumusers)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse '%s' as number of total users", strnumusers)
			}

		case regexTopic.MatchString(line):
			submatches := regexTopic.FindStringSubmatch(line)
			if len(submatches) != 3 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			topic = submatches[2]

			t, err = setClock(date, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

		case regexJoined.MatchString(line):
			submatches := regexJoined.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			t, err = setClock(date, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

			numusers++

		case regexQuit.MatchString(line):
			submatches := regexQuit.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			t, err = setClock(date, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

			numusers--

		case regexLeft.MatchString(line):
			submatches := regexLeft.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			t, err = setClock(date, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

			numusers--

		case regexNetsplitQuits.MatchString(line):
			submatches := regexNetsplitQuits.FindStringSubmatch(line)
			if len(submatches) != 3 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			quitsstr := submatches[2]

			numusers -= countNetsplitQuits(quitsstr)

			t, err = setClock(date, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

		case regexNetsplitJoins.MatchString(line):
			submatches := regexNetsplitJoins.FindStringSubmatch(line)
			if len(submatches) != 3 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			joinsstr := submatches[2]

			numusers += countNetsplitJoins(joinsstr)

			t, err = setClock(date, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

		// Must be last case
		case regexTimestamp.MatchString(line):
			submatches := regexTimestamp.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			t, err = setClock(date, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

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

func setClock(date time.Time, strclock string) (time.Time, error) {
	var hour, min int
	_, err := fmt.Sscanf(strclock, "%d:%d", &hour, &min)
	if err != nil {
		return time.Time{}, fmt.Errorf("Failed to scan '%s' as clock string: %s", strclock, err.Error())
	}
	t := date.Add(time.Hour * time.Duration(hour) + time.Minute * time.Duration(min))
	return t, nil
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

