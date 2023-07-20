package main

import (
	"fmt"
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"github.com/ivartj/norske-irc-kanaler.com/irc"
	"io"
	"os"
)

func mainUsage(w io.Writer) {
  fmt.Fprintf(w, "Usage: [-w <password>] [-U <username>] %s <server>[:<port>] \\#channel\n", os.Args[0])
}

func main() {
	var err error
	password := ""
	usePassword := false
	user := "testuser"
	nick := "testnick"
	tok := args.NewTokenizer(os.Args)
	positionals := []string{}
	for tok.Next() {
		if tok.IsOption() {
			switch tok.Arg() {
			case "-h", "--help":
				mainUsage(os.Stdout)
				return
			case "--version":
				fmt.Println("irc-cmd version 0.1.0")
				return
			case "-n":
				nick, err = tok.TakeParameter()
				if err != nil {
					panic(err)
				}
			case "-U":
				user, err = tok.TakeParameter()
				if err != nil {
					panic(err)
				}
			case "-w":
				usePassword = true
				password, err = tok.TakeParameter()
				if err != nil {
					panic(err)
				}
			}
		} else {
			positionals = append(positionals, tok.Arg())
		}
	}
	if len(positionals) != 2 {
		mainUsage(os.Stderr)
		os.Exit(1)
	}
	server := positionals[0]
	channel := positionals[1]

	var c *irc.Conn
	if usePassword {
		c, err = irc.ConnectWithPassword(server, nick, user, password, os.Stderr)
	} else {
		c, err = irc.Connect(server, nick, user, os.Stderr)
	}
	if err != nil {
		panic(err)
	}
	defer c.Disconnect()

	c.SendRawf("JOIN %s", channel)
	ev, err := c.WaitUntil("353") // RPL_NAMREPLY
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", ev.Args)
}
