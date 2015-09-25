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
	"database/sql"
)

var chanIllegalChars map[byte]string = map[byte]string{

// rfc2812 2.3.1 Message format in Augmented BNF
	'\x00' : "null-terminator",
	'\a' : "bjelletegnet",
	'\n' : "linjebrekk",
	'\r' : "linjeskift",
	' ' : "mellomrom",
	',' : "komma",
	':' : "kolon",
}
	
	
func chanCheckLoop() {
	for {
		chanCheckAll()
		time.Sleep(time.Hour * 25)
	}
}

func chanSuggestWebLink(name, server string) string {
	switch server {
	case "irc.freenode.net":
		return fmt.Sprintf("https://webchat.freenode.net/?channels=%s", url.QueryEscape(name))
	case "irc.efnet.pl":
		return fmt.Sprintf("http://chat.efnet.org:9090/?channels=%s&Login=Login", url.QueryEscape(name))
	default:
		return fmt.Sprintf("http://kiwiirc.com/client/%s/%s", url.QueryEscape(server), name)
	}
}

func chanStatus(ch *dbChannel) (string, bool) {
	status := ""
	switch {
	case !ch.checked:
		return "Ennå ikke undersøkt", true
	case ch.errmsg == "" && ch.numusers == 1:
		status = "1 bruker innlogget"
		status += " " + timeAgo(ch.lastcheck)
		return status, true
	case ch.errmsg == "":
		status = fmt.Sprintf("%d brukere innlogget", ch.numusers)
		status += " " + timeAgo(ch.lastcheck)
		return status, true
	default:
		status = "Feilmelding ved samling av informasjon: " + ch.errmsg + " (" + timeAgo(ch.lastcheck) + "), " + timeAgo(ch.lastcheck)
		return status, false
	}
}

func chanCheckAll() {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("Error occurred while checking channels: %s\n", err)
		}
	}()

	db := dbOpen()
	defer db.Close()
	servers := dbGetServers(db)
	chs, _ := dbGetApprovedChannels(db, 0, 9999)

	for _, server := range servers {
		chanCheckServer(db, server, chs)
	}
}

func chanCheckServer(db *sql.DB, server string, chs []dbChannel) {
	server_chs := []*dbChannel{}
	for _, ch := range chs {
		if ch.server != server {
			continue
		}

		if ch.checked {
			dur := time.Now().Sub(ch.lastcheck)
			if dur < time.Hour * 24 {
				continue
			}
		}

		server_chs = append(server_chs, &ch)
	}

	if len(server_chs) == 0 {
		return
	}

	bot, err := irc.Connect(server, conf.IRCBotNickname, conf.IRCBotRealname, nil)
	if err != nil {
		panic(err)
	}
	defer bot.Disconnect()

	for _, ch := range server_chs {
		if ch.server != server {
			continue
		}

		status, method, err := chanCheck(bot, ch.name)
		str := ""
		if err != nil {
			str = err.Error()
		}
		dbUpdateStatus(db, ch.name, ch.server, status.NumberOfUsers, status.Topic, method, str)
		time.Sleep(5 * time.Second)
	}

}

func chanCanonical(name, server string) (string, string) {
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

func chanValidate(name, server string) error {
	if !strings.HasPrefix(name, "#") {
		return errors.New("Kanalen må ha '#' prefiks")
	}

	for _, c := range []byte(name) {
		msg, ok := chanIllegalChars[c]
		if ok {
			return errors.New("Kanalnavnet kan ikke inneholde " + msg)
		}
	}

	if len(name) > 50 {
		return errors.New("Kanalnavnet kan ikke være lengre enn femti bytes (tegn, grovt oversatt)")
	}

	return nil
}

func chanCheck(bot *irc.Conn, name string) (*query.Result, string, error) {

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

