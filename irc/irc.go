package irc

import (
	"net"
	"fmt"
	"time"
	"bufio"
	"bytes"
	"strings"
	"errors"
	"io"
)

type HandleFunc func(*Conn, *Event)

var defaultHandlers map[string]HandleFunc = map[string]HandleFunc{
	"PING": pingHandler,
}

type Conn struct {
	net.Conn
	server string
	nick, user string
	scan *bufio.Scanner
	Events <-chan *Event
	events chan<- *Event
	handlers map[string]HandleFunc
	closed bool
}

func pingHandler(c *Conn, ev *Event) {
	reply := "PONG"
	for _, v := range ev.Args {
		reply += " " + v
	}
	c.SendRaw(reply)
}

func Connect(address, nick, user string) (*Conn, error) {
	// TODO: Do not add port if it is present in address
	sock, err := net.DialTimeout("tcp", address + ":6667", time.Second * 30)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to '%s': %s", address, err.Error())
	}

	scan := bufio.NewScanner(sock)
	scan.Split(bufio.ScanLines)

	ch := make(chan *Event)
	c := &Conn{sock, address, nick, user, scan, ch, ch, make(map[string]HandleFunc), false}
	c.SendRawf("NICK %s", nick)
	c.SendRawf("USER %s 0 * :%s", nick, user)

	go c.receiveMessages()
	_, err = c.WaitUntil("001") // Wait until welcome code

	return c, nil
}

func (c *Conn) GetServer() string {
	return c.server
}

func (c *Conn) handle(ev *Event) error {
	h, ok := c.handlers[ev.Code]
	if !ok {
		h, ok = defaultHandlers[ev.Code]
	}
	if !ok {
		if CodeIsError(ev.Code) {
			return errors.New(CodeString(ev.Code))
		}
		return nil
	}

	var err error
	err = nil
	defer func() {
		v := recover()
		err = fmt.Errorf("IRC server error: %v", v)
	}()

	h(c, ev)

	return err
}

func (c *Conn) SetHandler(code string, handler HandleFunc) {
	c.handlers[code] = handler
}

func (c *Conn) WaitUntil(codes ...string) (*Event, error) {
	var ev *Event
	for ev = range c.Events {
		for _, v := range codes {
			if ev.Code == v {
				goto ret
			}
		}
		if CodeIsError(ev.Code) {
			return nil, errors.New(CodeString(ev.Code))
		}
		err := c.handle(ev)
		if err != nil {
			return nil, err
		}
	}
	return nil, io.EOF

ret:
	return ev, nil
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

func (c *Conn) receiveMessages() {
	var err error
	const (
		stPrefix int	= iota
		stCode int	= iota
		stArgs int	= iota
	)

	for {
		ok := c.scan.Scan()
		if !ok {
			err = c.scan.Err()
			if err != nil && !c.closed {
				goto abort
			}
			goto eof
		}
		scanline := c.scan.Bytes()

		words := bufio.NewScanner(bytes.NewReader(scanline))
		words.Split(msgScan)

		state := stPrefix
		msg := &Event{}

		for words.Scan() {
			word := words.Text()

			switch state {
			case stPrefix:
				if strings.HasPrefix(word, ":") {
					msg.Prefix = word
					state = stCode
					break
				}
				fallthrough
			case stCode:
				msg.Code = word
				state = stArgs
			case stArgs:
				if strings.HasPrefix(word, ":") {
					word = strings.TrimPrefix(word, ":")
				}
				msg.Args = append(msg.Args, word)
			}
		}

		c.events <- msg
	}
eof:
	close(c.events)
	return
abort:
	panic(err)
}

func (c *Conn) SendRaw(msg string) {
	fmt.Fprint(c, msg)
	fmt.Fprint(c, "\r\n")
}

func (c *Conn) SendRawf(format string, args ...interface{}) {
	fmt.Fprintf(c, format, args...)
	fmt.Fprint(c, "\r\n")
}

func (c *Conn) Disconnect() {
	// TODO: Could probably be done more gracefully
	c.SendRaw("QUIT")
	c.Close()
	c.closed = true
}

type Event struct {
	Prefix string
	Code string
	Args []string
}

