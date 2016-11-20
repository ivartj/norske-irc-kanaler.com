package irclog

import (
	"bufio"
	"fmt"
	"regexp"
	"io"
	"strconv"
)

var (
	// 22:11 -!- Irssi: #example: Total of 113 nicks [6 ops, 0 halfops, 0 voices, 107 normal]
	regexTotalNick	= regexp.MustCompile(`^[0-9]{2}:[0-9]{2} -!- Irssi: #.+?: Total of ([0-9]+) nicks`)

	// 22:25 -!- FooNick [~BarUser@example-host] has joined #example
	regexJoined	= regexp.MustCompile(`^[0-9]{2}:[0-9]{2} -!- .+? \[~.+?@.+?\] has joined`)

	// 22:39 -!- FooNick [~BarUser@example-host] has quit [Ping timeout: 246 seconds]
	regexQuit	= regexp.MustCompile(`^[0-9]{2}:[0-9]{2} -!- .+? \[~.+?@.+?\] has quit`)

	// 18:16 -!- FooNick [~BarUser@example-host] has left #example [Leave message]
	regexLeft	= regexp.MustCompile(`^[0-9]{2}:[0-9]{2} -!- .+? \[~.+?@.+?\] has left`)

	// 09:34 -!- FooNick changed the topic of #example to: Lorem ipsum dolor sit amet
	regexTopic	= regexp.MustCompile(`^[0-9]{2}:[0-9]{2} -!- .+? changed the topic of .+? to: (.+)`)
)

func ChannelStatus(log io.Reader) (numusers int, topic string, err error) {

	scan := bufio.NewScanner(log)
	numusers = 0
	topic = ""
	err = nil

	for scan.Scan() {

		line := scan.Text()

		switch {

		case regexTotalNick.MatchString(line):

			submatches := regexTotalNick.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return 0, "", fmt.Errorf("Failed to capture by regex the number of total nicks from the line '%s'", line)
			}

			var err error
			numusers, err = strconv.Atoi(submatches[1])
			if err != nil {
				return 0, "", fmt.Errorf("Failed to parse '%s' as number of total users", submatches[1])
			}

		case regexTopic.MatchString(line):
			submatches := regexTopic.FindStringSubmatch(line)
			if len(submatches) != 2 {
				return 0, "", fmt.Errorf("Failed to capture by regex the topic from the line '%s'", line)
			}

			topic = submatches[1]

		case regexJoined.MatchString(line):
			numusers++

		case regexQuit.MatchString(line):
			numusers--

		case regexLeft.MatchString(line):
			numusers--

		}

	}

	if scan.Err() != nil {
		return 0, "", fmt.Errorf("Error on scanning line: %s", scan.Err().Error())
	}

	return numusers, topic, nil
}

