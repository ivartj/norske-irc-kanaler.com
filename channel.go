package main

import (
	"strings"
	"errors"
	"net/url"
	"fmt"
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
	New() bool
	Approved() bool
	ApproveTime() time.Time
	NumberOfUsers() int
	Topic() string
	Checked() bool
	CheckTime() time.Time
	Status() string
	Error() string
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

