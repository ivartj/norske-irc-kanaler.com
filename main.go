package main

import (
	"ivartj/args"
	"fmt"
	"io"
	"os"
	"path"
	"log"
)

var mainConfFilename string = ""

func mainUsage(out io.Writer, p *args.Parser) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  ircnorge CONFIGURATION-FILE")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Description:")
	fmt.Fprintln(out, "  Serves website that inspects and advertises")
	fmt.Fprintln(out, "  IRC channels.")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Options:")
	p.PrintUsage(out)
}

func mainArgs() {
	p := args.NewParser(os.Args[1:])
	p.AddOption('h', 'h', "help", false, "Prints help message")
	p.AddOption(301, '-', "version", false, "Prints version")

	for {
		code, arg := p.Parse() 

		if code == args.End {
			break
		}

		switch code {
		case args.Plain:
			mainConfFilename = arg
		case 'h':
			mainUsage(os.Stdout, p)
			os.Exit(0)
		case 301:
			fmt.Println("ircnorge version 0.1")
			os.Exit(0)
		case args.Unrecognized:
			fmt.Fprintf(os.Stderr, "Unrecognized option '%s'.\n", arg);
			os.Exit(1)
		case args.MissingOptionArgument:
			fmt.Fprintf(os.Stderr, "Missing option to '%s'.\n", arg);
			os.Exit(1)
		}
	}

	if mainConfFilename == "" {
		mainUsage(os.Stdout, p)
		os.Exit(1)
	}
}

func chdir() {
	dir := path.Dir(mainConfFilename)
	err := os.Chdir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to change to directory '%s': %s.\n", err.Error())
		os.Exit(1)
	}
}

func openlog() {
	if(conf.LogPath == "") {
		return
	}
	f, err := os.Create(conf.LogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file '%s': %s\n", conf.LogPath, err.Error())
		os.Exit(1)
	}
	log.SetOutput(f)
}

func main() {
	mainArgs()
	chdir()
	confParse(mainConfFilename)
	openlog()
	dbInit()
	go chanCheckLoop()
	serve()
}
