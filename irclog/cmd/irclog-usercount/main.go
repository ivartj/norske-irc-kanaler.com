package main

import (
	"github.com/ivartj/norske-irc-kanaler.com/irclog"
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"os"
	"fmt"
)

const (
	mainProgramName = "irclog-usercount"
	mainProgramVersion = "0.1-SNAPSHOT"
)

func mainArgs() ([]string) {

	tok := args.NewTokenizer(os.Args)
	logfiles := []string{}

	for tok.Next() {

		if tok.IsOption() {

			switch tok.Arg() {
			case "-h", "--help":
				fmt.Printf("Usage: %s logfile ...\n")
				os.Exit(0)

			case "--version":
				fmt.Printf("%s version %s\n", mainProgramName, mainProgramVersion)
				os.Exit(0)

			default:
				fmt.Fprintf(os.Stderr, "Unrecognized option, '%s'.\n", tok.Arg())
				os.Exit(1)

			}

		} else {
			logfiles = append(logfiles, tok.Arg())
		}

	}

	if tok.Err() != nil {
		fmt.Fprintf(os.Stderr, "Error occurred on processing command-line arguments: %s.\n", tok.Err().Error())
		os.Exit(1)
	}

	if len(logfiles) == 0 {
		fmt.Fprintf(os.Stderr, "No log files specified.\n")
		os.Exit(1)
	}

	return logfiles
}

func mainCount(logfilename string) error {

	logfile, err := os.Open(logfilename)
	if err != nil {
		return fmt.Errorf("Failed to open log file '%s': %s", logfilename, err.Error())
	}
	defer logfile.Close()

	numusers, topic, err := irclog.ChannelStatus(logfile)
	if err != nil {
		return fmt.Errorf("Failed to count number of users from '%s': %s", logfilename, err.Error())
	}

	fmt.Println(numusers, topic)

	return nil
}

func main() {
	logfilenames := mainArgs()
	for _, logfilename := range logfilenames {
		err := mainCount(logfilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s.\n", err.Error())
			os.Exit(1)
		}
	}
}

