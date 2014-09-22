package main

import (
	"ircnorge/irc"
	"strings"
	"errors"
	"fmt"
	"log"
	"time"
)

const (
	nickname string		= "ablegoyer"
)

func chanCheckLoop() {
	for {
		chanCheckAll()
		time.Sleep(time.Minute)
	}
}

func chanStatus(ch *dbChannel) string {
	status := ""
	switch {
	case !ch.checked:
		status = "Ennå ikke undersøkt"
	case ch.errmsg == "" && ch.numusers == 1:
		status = "1 bruker innlogget"
		status += " " + timeAgo(ch.lastcheck)
	case ch.errmsg == "":
		status = fmt.Sprintf("%d brukere innlogget", ch.numusers)
		status += " " + timeAgo(ch.lastcheck)
	default:
		status = "Feil: " + ch.errmsg + " (" + timeAgo(ch.lastcheck) + ")"
	}
	return status
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
			if dur < time.Hour * 24 * 7 {
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

	c, err := irc.Connect(server, nickname)
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
