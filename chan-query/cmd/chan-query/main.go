package main

import (
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"github.com/ivartj/norske-irc-kanaler.com/irc"
	query "github.com/ivartj/norske-irc-kanaler.com/chan-query"
	"os"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	mainName		= "chan-query"
	mainVersion		= "1.0"
)

func mainPrintUsage(w io.Writer) {
	fmt.Fprintf(w, `Usage:
  %s [OPTIONS] #channel@irc.example.com ...

Note:
  If using a conventional Unix shell, you should escape the number sign (#).

Options:
  -h, --help             Prints help message.
  --version              Prints version.
  -t, --timeout=SECONDS  Timeout between checking channels.
                         Default is 2 seconds.
  -m, --method=METHOD    Method to use to query channel.
                         If used multiple times the methods will be tried in
                         sequence.
                         The default is to only use the 'list' method.

`, mainName)

	fmt.Fprintf(w, `Methods:
  The following methods can be specified through the --method option.

`)
	for _, v := range query.GetMethods() {
		fmt.Fprintf(w, "  %s", v.Name())
		for i := 2 + len(v.Name()); i < len("  -h, --help             "); i++ {
			fmt.Fprintf(w, " ")
		}
		fmt.Fprintf(w, "%s\n", v.Description())
	}

	fmt.Fprintln(w, "")
}

var (
	confChannels = make(map[string][]string)
	confNickname = "norbot5123"
	confUsername = "norbot"
	confTimeout = time.Second * 2
	confQueryMethods = []*query.Method {}
)

func mainParseArgs() {
	defer func() {
		err, isErr := recover().(error)
		if isErr {
			fmt.Fprintf(os.Stderr, "Error on processing command-line arguments: %s.\n", err.Error())
			os.Exit(1)
		}
	}()

	tok := args.NewTokenizer(os.Args)

	channelStrings := []string{}

	for tok.Next() {
		arg := tok.Arg()
		isOption := tok.IsOption()

		if !isOption {
			channelStrings = append(channelStrings, arg)
			continue
		}

		// TODO: Make username and nickname options

		switch arg {
		case "-h", "--help":
			mainPrintUsage(os.Stdout)
			os.Exit(0)
		case "--version":
			fmt.Printf("%s version %s\n", mainName, mainVersion)
			os.Exit(0)
		case "-t", "--timeout":
			param, err := tok.TakeParameter()
			if err != nil {
				panic(err)
			}
			secs, err := strconv.ParseFloat(param, 64)
			if err != nil {
				panic(fmt.Errorf("Failed to parse parameter to %s, %s", arg, param))
			}
			confTimeout = time.Duration(float64(time.Second) * secs)
		case "-m", "--query-method":
			param, err := tok.TakeParameter()
			if err != nil {
				panic(err)
			}
			m, ok := query.GetMethodByName(param)
			if !ok {
				panic(fmt.Errorf("Not a recognized query method, '%s'", param))
			}
			confQueryMethods = append(confQueryMethods, m)
		default:
			panic(fmt.Errorf("Unexpected option, '%s'", arg))
		}
	}

	if tok.Err() != nil {
		panic(tok.Err())
	}

	for _, v := range channelStrings {
		div := strings.Split(v, "@")
		if len(div) != 2 {
			panic(fmt.Errorf("Invalid channel string '%s'", v))
		}

		// TODO: Validate and canonicalize channel and server names
		server := div[1]
		channel := div[0]

		serverChannels, ok := confChannels[server]
		if !ok {
			serverChannels = []string{}
		}
		serverChannels = append(serverChannels, channel)
		confChannels[server] = serverChannels
	}

	if len(confQueryMethods) == 0 {
		confQueryMethods = append(confQueryMethods, query.ListMethod)
	}
}

func queryChannelsOnServer(server string, channels []string) {
	defer func() {
		err, isErr := recover().(error)
		if isErr {
			fmt.Fprintf(os.Stderr, "Error occurred when querying channels on %s: %s.\n", server, err.Error())
		}
	}()

	bot, err := irc.Connect(server, confNickname, confUsername)
	if err != nil {
		panic(err)
	}
	defer bot.Disconnect()

	for i, ch := range channels {

		errs := []error{}

		for _, method := range confQueryMethods {
			res, err := method.Query(bot, ch)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			fmt.Printf("%s\t%d\t%s\t%s\n", res.Name, res.NumberOfUsers, method.Name(), res.Topic)
			goto next
		}

		fmt.Fprintf(os.Stderr, "Failed to query %s@%s,\n", ch, server)
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "  %s\n", err.Error())
		}

next:
		if i + 1 != len(channels) {
			time.Sleep(confTimeout)
		}
	}

}

func main() {
	mainParseArgs()
	for server, channels := range confChannels {
		queryChannelsOnServer(server, channels)
	}
}

