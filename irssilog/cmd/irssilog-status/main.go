package main

import (
	irclog "github.com/ivartj/norske-irc-kanaler.com/irssilog"
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"os"
	"fmt"
	"strings"
)

const (
	mainProgramName = "irssilog-status"
	mainProgramVersion = "0.1-SNAPSHOT"
)

type mainChannel struct{
	name, network string
}

// key is network and value is names
var mainConfNetworkNames = map[string][]string{}
var mainConfLogDirectory = "."
var mainConfChannels = []mainChannel{}

func mainArgs() {

	tok := args.NewTokenizer(os.Args)

	for tok.Next() {

		if tok.IsOption() {

			switch tok.Arg() {
			case "-h", "--help":
				fmt.Printf("Usage: %s logfile ...\n")
				os.Exit(0)

			case "--version":
				fmt.Printf("%s version %s\n", mainProgramName, mainProgramVersion)
				os.Exit(0)

			case "-d":
				param, err := tok.TakeParameter()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to take parameter to '%s' option: %S", tok.Arg(), err.Error())
					os.Exit(1)
				}

				mainConfLogDirectory = param

			case "-a":
				param, err := tok.TakeParameter()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed take parameter to '%s' option: %s", tok.Arg(), err.Error())
					os.Exit(1)
				}

				namenetwork := strings.Split(param, ":")
				if len(namenetwork) != 2 {
					fmt.Fprintf(os.Stderr, "'%s' is not a valid name-network association.\n", param)
					os.Exit(1)
				}
				name := namenetwork[0]
				network := namenetwork[1]

				names, ok := mainConfNetworkNames[network]
				if !ok {
					names = []string{}
				}
				mainConfNetworkNames[network] = append(names, name)

			default:
				fmt.Fprintf(os.Stderr, "Unrecognized option, '%s'.\n", tok.Arg())
				os.Exit(1)

			}

		} else {
			namenetwork := strings.Split(tok.Arg(), "@")
			if len(namenetwork) != 2 {
				fmt.Fprintf(os.Stderr, "'%s' is not a valid channel", tok.Arg())
				os.Exit(1)
			}
			name := namenetwork[0]
			network := namenetwork[1]
			mainConfChannels = append(mainConfChannels, mainChannel{name: name, network: network}) 
		}

	}

	if tok.Err() != nil {
		fmt.Fprintf(os.Stderr, "Error occurred on processing command-line arguments: %s.\n", tok.Err().Error())
		os.Exit(1)
	}

	if len(mainConfChannels) == 0 {
		fmt.Fprintf(os.Stderr, "No channels specified.\n")
		os.Exit(1)
	}
}

func main() {
	mainArgs()
	logctx := irclog.New(mainConfLogDirectory, mainConfNetworkNames)
	for _, channel := range mainConfChannels {
		cs, err := logctx.ChannelStatus(channel.name, channel.network)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s.\n", err.Error())
			os.Exit(1)
		}
		fmt.Println(cs.NumUsers, cs.Topic)
	}
}

