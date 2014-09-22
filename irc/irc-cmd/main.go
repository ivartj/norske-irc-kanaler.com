package main

import (
	"ircnorge/irc"
	"fmt"
)

func main() {
	c, err := irc.Connect("irc.freenode.net", "ablegoyer")
	if err != nil {
		panic(err)
	}
	
	ch, err := c.Join("##fest")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", ch.Names)
}
