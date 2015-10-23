package main

import (
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"fmt"
	"io"
	"os"
	"path"
	"log"
)

var mainConfFilename string = ""

const (
	mainName		= "norske-irc-kanaler.com"
	mainVersion		= "1.0"
)

func mainUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintf(out,  "  %s CONFIGURATION_FILE\n", mainName)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Description:")
	fmt.Fprintln(out, "  Serves website that inspects and advertises")
	fmt.Fprintln(out, "  IRC channels.")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "  -h, --help  Prints help message")
	fmt.Fprintln(out, "  --version   Prints version")
	fmt.Fprintln(out, "")
}

func mainArgs() {
	tok := args.NewTokenizer(os.Args)

	for tok.Next() {
		arg := tok.Arg()
		isOption := tok.IsOption()

		switch {
		case isOption:
			switch arg {
			case "-h", "--help":
				mainUsage(os.Stdout)
				os.Exit(0)
			case "--version":
				fmt.Printf("%s version %s\n", mainName, mainVersion)
				os.Exit(0)
			}
		case !isOption:
			if mainConfFilename != "" {
				mainUsage(os.Stderr)
				os.Exit(1)
			}
			mainConfFilename = arg
		}
	}

	if tok.Err() != nil {
		fmt.Fprintf(os.Stderr, "Error on processing arguments: %s.\n", tok.Err().Error())
		os.Exit(1)
	}

	if mainConfFilename == "" {
		mainUsage(os.Stderr)
		os.Exit(1)
	}
}

func mainChangeDirectory() {
	dir := path.Dir(mainConfFilename)
	err := os.Chdir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to change to directory '%s': %s.\n", dir, err.Error())
		os.Exit(1)
	}
}

func mainOpenLog() {
	if(conf.LogPath == "") {
		return
	}
	f, err := os.Create(conf.LogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file '%s': %s\n", conf.LogPath, err.Error())
		os.Exit(1)
	}
	mw := io.MultiWriter(f, os.Stderr)
	log.SetOutput(mw)
}

func main() {
	mainArgs()
	mainChangeDirectory()
	confParse(mainConfFilename)
	mainOpenLog()
	dbInit()
	if conf.IRCBotEnable {
		go channelCheckLoop()
	}
	serve()
}

