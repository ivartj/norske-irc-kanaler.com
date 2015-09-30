package main

import (
	"github.com/ivartj/norske-irc-kanaler.com/irc"
	"github.com/ivartj/norske-irc-kanaler.com/chan-query"
	"strings"
	"errors"
	"net/url"
	"bytes"
	"fmt"
	"log"
	"time"
)

var channelIllegalChars map[byte]string = map[byte]string{

// rfc2812 2.3.1 Message format in Augmented BNF
	'\x00' : "null-terminator",
	'\a' : "bjelletegnet",
	'\n' : "linjebrekk",
	'\r' : "linjeskift",
	' ' : "mellomrom",
	',' : "komma",
	':' : "kolon",
}

type channel interface {
	Name() string
	Network() string
	Weblink() string
	Description() string
	SubmitTime() time.Time
	Approved() bool
	ApproveTime() time.Time
	NumberOfUsers() int
	Topic() string
	Checked() bool
	CheckTime() time.Time
	Error() string
}

func channelCheckLoop() {
	for {
		channelCheckAll()
		time.Sleep(time.Hour * 25)
	}
}

func channelSuggestWebLink(name, server string) string {
	switch server {
	case "irc.freenode.net":
		return fmt.Sprintf("https://webchat.freenode.net/?channels=%s", url.QueryEscape(name))
	case "irc.efnet.pl":
		return fmt.Sprintf("http://chat.efnet.org:9090/?channels=%s&Login=Login", url.QueryEscape(name))
	default:
		return fmt.Sprintf("http://kiwiirc.com/client/%s/%s", url.QueryEscape(server), name)
	}
}

func channelStatusString(ch channel) (string, bool) {
	status := ""
	switch {
	case !ch.Checked():
		return "Ennå ikke undersøkt", true
	case ch.Error() == "" && ch.NumberOfUsers() == 1:
		status = "1 bruker innlogget"
		status += " " + timeAgo(ch.CheckTime())
		return status, true
	case ch.Error() == "":
		status = fmt.Sprintf("%d brukere innlogget", ch.NumberOfUsers())
		status += " " + timeAgo(ch.CheckTime())
		return status, true
	default:
		status = "Feilmelding ved samling av informasjon: " + ch.Error() + " (" + timeAgo(ch.CheckTime()) + ")"
		return status, false
	}
}

func channelCheckAll() {
	defer func() {
		err, isErr := recover().(error)
		if isErr {
			log.Printf("Error occurred while checking channels: %s\n", err.Error())
		}
	}()

	db, err := dbOpen()
	if err != nil {
		log.Fatalf("Failed to open database: %s.\n", err.Error())
	}
	defer db.Close()

	networks, err := db.GetNetworks()
	if err != nil {
		log.Fatalf("Failed to get network data from database: %s", err.Error())
	}
	chs, err := db.GetApprovedChannels(0, 9999)
	if err != nil {
		log.Fatalf("Failed to get list of approved channel from database: %s", err.Error())
	}

	for _, network := range networks {
		channelCheckServer(db, network, chs)
	}
}

func channelCheckServer(db *dbConn, network *dbNetwork, chs []channel) {
	defer func() {
		err, isErr := recover().(error)
		if isErr {
			log.Printf("Error occurred while checking channels on '%s': %s\n", network.network, err.Error())
		}
	}()

	network_chs := []channel{}
	for _, ch := range chs {
		if ch.Network() != network.network {
			continue
		}

		if ch.Checked() {
			dur := time.Now().Sub(ch.CheckTime())
			if dur < time.Hour * 24 {
				continue
			}
		}

		network_chs = append(network_chs, ch)
	}

	if len(network_chs) == 0 {
		return
	}

	log.Printf("Checking channels on %s.\n", network.network)

	var bot *irc.Conn
	var err error
	for _, server := range network.servers {
		bot, err = irc.Connect(server, conf.IRCBotNickname, conf.IRCBotRealname, nil)
		if err != nil {
			log.Printf("Failed to connect to %s: %s.\n", server, err.Error())
			continue
		}
		defer bot.Disconnect()
		break
	}
	if bot == nil {
		log.Printf("Could not connect to any address associated with %s.\n", network.network)
		return
	}

	for _, ch := range network_chs {
		if ch.Network() != network.network {
			continue
		}

		log.Printf("Checking %s@%s.\n", ch.Name(), ch.Network())
		status, method, err := channelCheck(bot, ch.Name())
		str := ""
		if err != nil {
			str = err.Error()
			err = db.UpdateStatus(ch.Name(), ch.Network(), 0, "", "fail", str)
		} else {
			log.Printf("%s@%s %d Topic: %s\n", ch.Name(), ch.Network(), status.NumberOfUsers, status.Topic)
			err = db.UpdateStatus(ch.Name(), ch.Network(), status.NumberOfUsers, status.Topic, method, str)
		}
		if err != nil {
			log.Fatalf("Database error when updating channel status: %s.\n", err.Error())
		}
		time.Sleep(5 * time.Second)
	}

}

func channelAddressCanonical(name, server string) (string, string) {
	name = strings.TrimLeft(name, " \t\r\n")
	name = strings.TrimRight(name, " \t\r\n")
	server = strings.TrimLeft(server, " \t\r\n")
	server = strings.TrimRight(server, " ")

	server = strings.TrimRight(server, ".")

	name = strings.ToLower(name)
	server = strings.ToLower(server)

	// rfc2812 2.2 Character codes
	name = strings.Replace(name, "[", "{", -1)
	name = strings.Replace(name, "]", "}", -1)
	name = strings.Replace(name, "\\", "|", -1)
	name = strings.Replace(name, "^", "~", -1)

	return name, server
}

func channelAddressValidate(name, server string) error {
	if !strings.HasPrefix(name, "#") {
		return errors.New("Kanalen må ha '#' prefiks")
	}

	for _, c := range []byte(name) {
		msg, ok := channelIllegalChars[c]
		if ok {
			return errors.New("Kanalnavnet kan ikke inneholde " + msg)
		}
	}

	if len(name) > 50 {
		return errors.New("Kanalnavnet kan ikke være lengre enn femti bytes (tegn, grovt oversatt)")
	}

	return nil
}

func channelCheck(bot *irc.Conn, name string) (*query.Result, string, error) {

	log := bytes.NewBuffer([]byte{})

	for _, method := range query.GetMethods() {
		res, err := method.Query(bot, name)
		if err == nil {
			return res, method.Name(), nil
		}
		fmt.Fprintf(log, "method %s failed: %s\n", method.Name(), err.Error())
	}

	return nil, "", errors.New(log.String())
}

