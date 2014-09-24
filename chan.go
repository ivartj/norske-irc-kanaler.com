package main

import (
	"ircnorge/irc"
	"strings"
	"errors"
	"net/url"
	"fmt"
	"log"
	"time"
)

func chanCheckLoop() {
	for {
		chanCheckAll()
		time.Sleep(time.Minute)
	}
}

func chanSuggestWebLink(name, server string) string {
	switch server {
	case "irc.freenode.net":
		return fmt.Sprintf("https://webchat.freenode.net/?channels=%s", url.QueryEscape(name))
	case "irc.efnet.pl":
		return fmt.Sprintf("http://chat.efnet.org:9090/?channels=%s&Login=Login", url.QueryEscape(name))
	default:
		return fmt.Sprintf("http://kiwiirc.com/%s/%s", url.QueryEscape(server), name)
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

	c := dbOpen()
	defer c.Close()
	chs, _ := dbGetApprovedChannels(c, 0, 9999)

	for _, v := range chs {
		if v.checked {
			dur := time.Now().Sub(v.lastcheck)
			if dur < time.Hour * 24 {
				continue
			}
		}
		n, err := chanCheck(v.name, v.server)
		str := ""
		if err != nil {
			str = err.Error()
		}
		dbUpdateStatus(c, v.name, v.server, n, str)
		time.Sleep(time.Second * 5)
	}
}

func chanValidate(name, server string) error {
	if !strings.HasPrefix(name, "#") {
		return errors.New("Kanalen må ha '#' prefiks")
	}

	if strings.ContainsRune(name, ' ') {
		return errors.New("Kanalnavnet kan ikke ha mellomrom")
	}

	if len(name) > 50 {
		return errors.New("Kanalnavnet kan ikke være lengre enn femti bytes (tegn, grovt oversatt)")
	}

	return nil
}

func chanCheck(name, server string) (int, error) {

	c, err := irc.Connect(server, conf.IRCBotNickname, conf.IRCBotRealname)
	if err != nil {
		return 0, err
	}
	defer c.Quit()

	ch, err := c.Join(name)
	if err != nil {
		return 0, err
	}

	return len(ch.Names) - 1, nil
}
