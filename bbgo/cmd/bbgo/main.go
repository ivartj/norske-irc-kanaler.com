package main

import (
	"fmt"
	"github.com/ivartj/norske-irc-kanaler.com/args"
	"github.com/ivartj/norske-irc-kanaler.com/bbgo"
	"io"
	"os"
)

const (
	mainName    = "bbgo"
	mainVersion = "1.0"
)

var (
	mainInput  io.Reader = os.Stdin
	mainOutput io.Writer = os.Stdout
)

func mainUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintf(w, "  %s [ -o OUTPUT ] [ INPUT ]\n", mainName)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Description:")
	fmt.Fprintln(w, "  Processes BBCode and outputs HTML.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  -h, --help             Prints help message.")
	fmt.Fprintln(w, "  --version              Prints version.")
	fmt.Fprintln(w, "  -o, --output=FILENAME  Specifies output filename.")
	fmt.Fprintln(w, "")
}

func mainParseArguments(argv []string) {
	tok := args.NewTokenizer(argv)

	for tok.Next() {
		if !tok.IsOption() {
			if conf.inputFilename.isSet() {
				fmt.Fprintf(os.Stderr, "Can't specify more than one input file.\n")
				os.Exit(1)
			}
			conf.inputFilename.set(tok.Arg())
		}

		switch tok.Arg() {
		case "-h":
			fallthrough
		case "--help":
			mainUsage(os.Stdout)
			os.Exit(0)

		case "--version":
			fmt.Printf("%s version %s\n", mainName, mainVersion)
			os.Exit(0)

		case "-o":
			fallthrough
		case "--output":
			if conf.outputFilename.isSet() {
				fmt.Fprintf(os.Stderr, "Can't specify more than one output file.\n")
				os.Exit(1)
			}
			param, err := tok.TakeParameter()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error on parsing command-line arguments: %s.\n", err.Error())
				os.Exit(1)
			}
			conf.outputFilename.set(param)
		}
	}

	if tok.Err() != nil {
		fmt.Fprintf(os.Stderr, "Error on parsing command-line arguments: %s.\n", tok.Err().Error())
		os.Exit(1)
	}
}

func mainOpenFiles() {

	var err error

	if conf.inputFilename.isSet() {
		mainInput, err = os.Open(conf.inputFilename.get())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open '%s': %s.\n", conf.inputFilename.get(), err.Error())
			os.Exit(1)
		}
	}

	if conf.outputFilename.isSet() {
		mainInput, err = os.Create(conf.outputFilename.get())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open '%s': %s.\n", conf.outputFilename.get(), err.Error())
			os.Exit(1)
		}
	}
}

func main() {
	mainParseArguments(os.Args)
	mainOpenFiles()
	err := bbgo.Process(mainInput, mainOutput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error on processing bbcode file: %s.\n", err.Error())
		os.Exit(1)
	}
}
