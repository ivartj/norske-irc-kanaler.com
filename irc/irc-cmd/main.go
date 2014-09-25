package main

import (
	"ircnorge/irc"
	"fmt"
)

func main() {
	c, err := irc.Connect("irc.efnet.pl", "asbestos12345", "asbestos12345")
	if err != nil {
		panic(err)
	}
	defer c.Quit()

	ch, err := c.Join("#hardware.no")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", ch.Names)
}
