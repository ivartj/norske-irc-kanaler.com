package irssilog

import (
	"bufio"
	"fmt"
	"regexp"
	"io"
	"strconv"
	"time"
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
	timeFormatClock = "15:04"

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
)


func GetChannelStatusFromLog(log io.Reader) (ChannelStatus, error) {

	scan := bufio.NewScanner(log)
	numusers := 0
	topic := ""
	t := time.Time{}

	dateSet := true

	for scan.Scan() {

		line := scan.Text()

		switch {

		case regexLogOpened.MatchString(line):

			submatches := regexLogOpened.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Failed to capture by regex the date from the line '%s'", line)
			}

			// TODO: Consider that the log is from a system with different timezone
			var err error
			t, err = time.Parse(time.ANSIC, submatches[1])
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse the time given in '%s': %s", line, err.Error())
			}

			dateSet = true

		case regexDayChanged.MatchString(line):

			submatches := regexDayChanged.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Failed to capture by regex the date given in the line '%s'", line)
			}

			// TODO: Consider that the log is from a system with different timezone
			var err error
			t, err = time.Parse(timeFormatDayChanged, submatches[1])
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse the time given in '%s': %s", line, err.Error())
			}

			dateSet = true

		case regexTotalNick.MatchString(line):

			submatches := regexTotalNick.FindStringSubmatch(line)
			if len(submatches) != 3 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			strnumusers := submatches[2]

			err := setClock(&t, strclock)
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

			err := setClock(&t, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

		case regexJoined.MatchString(line):
			submatches := regexJoined.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return ChannelStatus{}, fmt.Errorf("Unexpected number of submatches in the line '%s'", line)
			}
			strclock := submatches[1]
			err := setClock(&t, strclock)
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
			err := setClock(&t, strclock)
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
			err := setClock(&t, strclock)
			if err != nil {
				return ChannelStatus{}, fmt.Errorf("Failed to parse timestamp on line '%s': %s", line, err.Error())
			}

			numusers--

		}

	}

	if scan.Err() != nil {
		return ChannelStatus{}, fmt.Errorf("Error on scanning line: %s", scan.Err().Error())
	}

	if !dateSet {
		return ChannelStatus{}, fmt.Errorf("No date was found throughout the log")
	}

	return ChannelStatus{Time: t, NumUsers: numusers, Topic: topic}, nil
}

func setClock(t *time.Time, strclock string) error {
	var hour, min int
	_, err := fmt.Sscanf(strclock, "%d:%d", &hour, &min)
	if err != nil {
		return fmt.Errorf("Failed to scan '%s' as clock string: %s", strclock, err.Error())
	}
	*t = t.Truncate(24 * time.Hour)
	*t = t.Add(time.Hour * time.Duration(hour) + time.Minute * time.Duration(min))
	return nil
}

