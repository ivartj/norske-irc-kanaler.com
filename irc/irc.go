package irc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

type HandleFunc func(*Conn, *Event)

var defaultHandlers map[string]HandleFunc = map[string]HandleFunc{
	"PING": pingHandler,
}

type Conn struct {
	net.Conn
	irclog      io.Writer
	irclogMutex sync.Mutex
	server      string
	nick, user  string
	scan        *bufio.Scanner
	Events      <-chan *Event
	events      chan<- *Event
	handlers    map[string]HandleFunc
	closed      bool
	timeout     time.Duration
}

type Config struct {
	Nick     string
	User     string
	Password string    // optional
	Log      io.Writer // optional
}

func pingHandler(c *Conn, ev *Event) {
	reply := "PONG"
	for _, v := range ev.Args {
		reply += " " + v
	}
	c.SendRaw(reply)
}

var addressWithPort = regexp.MustCompile("[a-zA-Z0-9\\.-]+:[0-9]+")

func New(conn net.Conn, config *Config) (*Conn, error) {
	if config.Nick == "" {
		return nil, fmt.Errorf("no nick specified for the IRC session")
	}
	if config.User == "" {
		return nil, fmt.Errorf("no user specified for the IRC session")
	}
	log := config.Log
	if log == nil {
		log = io.Discard
	}
	usePassword := config.Password != ""
	return connect(conn, conn.RemoteAddr().String(), config.Nick, config.User, usePassword, config.Password, log)
}

func getSocket(address string) (net.Conn, error) {
	if !addressWithPort.MatchString(address) {
		address = address + ":6667"
	}

	sock, err := net.DialTimeout("tcp", address, time.Second*30)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to '%s': %s", address, err.Error())
	}

	return sock, nil
}

func Connect(address, nick, user string, irclog io.Writer) (*Conn, error) {
	sock, err := getSocket(address)
	if err != nil {
		return nil, err
	}
	return connect(sock, address, nick, user, false, "", irclog)
}

func ConnectWithPassword(address, nick, user, password string, irclog io.Writer) (*Conn, error) {
	sock, err := getSocket(address)
	if err != nil {
		return nil, err
	}
	return connect(sock, address, nick, user, true, password, irclog)
}

func connect(sock net.Conn, address, nick, user string, usePassword bool, password string, irclog io.Writer) (*Conn, error) {
	scan := bufio.NewScanner(sock)
	scan.Split(bufio.ScanLines)

	ch := make(chan *Event)
	c := &Conn{
		Conn:        sock,
		irclog:      irclog,
		irclogMutex: sync.Mutex{},
		server:      address,
		nick:        nick,
		user:        user,
		scan:        scan,
		Events:      ch,
		events:      ch,
		handlers:    make(map[string]HandleFunc),
		closed:      false,
		timeout:     time.Second * 10,
	}
	if usePassword {
		c.SendRawf("PASS %s", password)
	}
	c.SendRawf("NICK %s", nick)
	c.SendRawf("USER %s 0 * :%s", user, nick)

	go c.receiveMessages()
	_, err := c.WaitUntil("001") // Wait until welcome code
	if err != nil {
		c.Disconnect()
		return nil, err
	}

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
	timeout := time.After(c.timeout)
	for {
		select {
		case <-timeout:
			return nil, errors.New("timeout")
		case ev, ok := <-c.Events:
			if !ok {
				return nil, io.EOF
			}
			for _, v := range codes {
				if ev.Code == v {
					return ev, nil
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
	}
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

	for end = start + 1; end < len(data) && data[end] != ' '; end++ {
	}

	return end, data[start:end], nil
}

func (c *Conn) receiveMessages() {
	var err error
	const (
		stPrefix int = iota
		stCode   int = iota
		stArgs   int = iota
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
		if c.irclog != nil {
			fmt.Fprintln(c.irclog, "<==", string(scanline))
		}

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
	c.SendRawf("%s", msg)
}

func (c *Conn) SendRawf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	fmt.Fprintf(c, "%s\r\n", s)
	if c.irclog != nil {
		fmt.Fprintf(c.irclog, "==> %s\n", s)
	}
}

func (c *Conn) Disconnect() {
	// TODO: Could probably be done more gracefully
	c.SendRaw("QUIT")
	c.closed = true // this line has to before c.Close(), or else we have to put a mutex around these statements
	c.Close()
}

type Event struct {
	Prefix string
	Code   string
	Args   []string
}
