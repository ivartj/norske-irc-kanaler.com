package irc

import (
	"net"
	"fmt"
	"io"
	"bufio"
	"errors"
	"bytes"
	"strings"
	"time"
)

type Conn struct{
	sock net.Conn
	nick string
	scan *bufio.Scanner
}

func Connect(server, nick string) (*Conn, error) {
	sock, err := net.DialTimeout("tcp", server + ":6667", time.Second * 30)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to '%s': %s", server, err.Error())
	}

	scan := bufio.NewScanner(sock)
	scan.Split(bufio.ScanLines)

	c := &Conn{sock, nick, scan}

	fmt.Fprintf(sock, "NICK %s\r\n", nick)
	fmt.Fprintf(sock, "USER %s 0 * :IRC-chat Norge (www.ircnorge.org)\r\n", nick)

	sock.SetReadDeadline(time.Now().Add(time.Minute))

	msg, err := c.NextMessage()

	for ; err == nil; msg, err = c.NextMessage() {
		if msg.Command == "001" {
			goto out
		}
		if msg.Command == "PING" {
			pong := "PONG"
			for _, v := range msg.Args {
				pong += " " + v
			}
			fmt.Fprintf(c.sock, "%s\r\n", pong)
		}
		if CodeIsError(msg.Command) {
			err = errors.New(CodeString(msg.Command))
			goto out
		}

	}
out:

	if err != nil {
		sock.Close()
		return nil, fmt.Errorf("Error upon connecting to server '%s': %s", server, err.Error())
	}

	sock.SetReadDeadline(time.Time{})

	return c, nil
}

type Channel struct {
	Names []string
}

func (c *Conn) Join(channel string) (*Channel, error) {

	fmt.Fprintf(c.sock, "JOIN %s\r\n", channel)

	ch := &Channel{
		Names: []string{},
	}

	c.sock.SetReadDeadline(time.Now().Add(time.Minute))

	msg, err := c.NextMessage()
	for ; err == nil; msg, err = c.NextMessage() {

		// RPL_NAMREPLY
		if msg.Command == "353" {
			if len(msg.Args) != 0 {
				ch.Names = strings.Split(msg.Args[len(msg.Args) - 1], " ")
			}
			goto out
		}

		if msg.Command == "PING" {
			pong := "PONG"
			for _, v := range msg.Args {
				pong += " " + v
			}
			fmt.Fprintf(c.sock, "%s\r\n", pong)
		}
		if CodeIsError(msg.Command) {
			err = errors.New(CodeString(msg.Command))
			goto out
		}

	}
out:

	if err != nil {
		return nil, fmt.Errorf("Error upon joining channel '%s': %s", channel, err.Error())
	}

	c.sock.SetReadDeadline(time.Time{})

	return ch, nil
}

func (c *Conn) Quit() {
	fmt.Fprintf(c.sock, "QUIT\r\n")
	c.sock.Close()
}

type Message struct {
	Prefix string
	Command string
	Args []string
}

func msgScan(data []byte, atEOF bool) (advance int, token []byte, err error) {

	if len(data) == 0 {
		return 0, nil, nil
	}

	start := 1
	end := 0

	// prefix
	if data[0] != ' ' {
		start = 0
	}

	// trail
	if len(data) >= 2 && data[1] == ':' {
		start = 2
		if !atEOF {
			return 0, nil, nil
		}
		end = len(data)
		return len(data), data[start:end], nil
	}

	for end = start + 1; end < len(data) && data[end] != ' '; end++ {}

	return end, data[start:end], nil
}

func (c *Conn) NextMessage() (*Message, error) {

	const (
		stPrefix int	= iota
		stCommand int	= iota
		stArgs int	= iota
	)

	ok := c.scan.Scan()
	if !ok {
		err := c.scan.Err()
		if err == nil {
			err = io.EOF
		}
		return nil, err
	}
	scanline := c.scan.Bytes()

	words := bufio.NewScanner(bytes.NewReader(scanline))
	words.Split(msgScan)

	state := stPrefix
	msg := &Message{}

	for words.Scan() {
		word := words.Text()

		switch state {
		case stPrefix:
			if strings.HasPrefix(word, ":") {
				msg.Prefix = word
				state = stCommand
				break
			}
			fallthrough
		case stCommand:
			msg.Command = word
			state = stArgs
		case stArgs:
			if strings.HasPrefix(word, ":") {
				word = strings.TrimPrefix(word, ":")
			}
			msg.Args = append(msg.Args, word)
		}
	}

	return msg, nil
}
