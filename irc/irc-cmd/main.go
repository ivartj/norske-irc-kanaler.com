package main

import (
	"ircnorge/irc"
	"fmt"
)

func main() {
	c, err := irc.Connect("irc.efnet.pl", "asbestos12345")
	if err != nil {
		panic(err)
	}
	defer c.Quit()

	fmt.Println(".")
	
	ch, err := c.Join("#itpro")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", ch.Names)
}
