package main

import (
	"ircnorge/irc"
	"fmt"
)

func main() {
	c, err := irc.Connect("irc.undernet.org", "asbestos12345")
	if err != nil {
		panic(err)
	}
	defer c.Quit()

	ch, err := c.Join("#norge")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", ch.Names)
}
