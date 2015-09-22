package args

import (
	"io"
	"testing"
)

func TestTokenizer(t *testing.T) {
	argv := []string{ "-h", "--version", "-xzvf", "--output=filename", "-i=input" }
	lex := NewTokenizer(argv)

	var (
		optHelp, optVersion bool
		optX, optZ, optV, optF bool
		optOutput, optInput bool
	)

	for {
		arg, err := lex.Next()
		if err != nil && err != io.EOF {
			t.Fatalf("Unexpected error returned from Tokenizer.next(): %s\n", err.Error())
		}
		if err != nil {
			goto loopExit
		}

		switch arg {
		case "-h", "--help":
			optHelp = true
		case "--version":
			optVersion = true
		case "-x": optX = true
		case "-z": optZ = true
		case "-v": optV = true
		case "-f": optF = true
		case "-o", "--output":
			optOutput = true
			param, err := lex.TakeParameter()
			if err != nil {
				t.Fatalf("Error on calling Tokenizer.TakeParameter: %s\n", err.Error())
			}
			if param != "filename" {
				t.Fatalf("'%s' does not match the expected parameter 'filename'.\n", param)
			}
		case "-i", "--input":
			optInput = true
			param, err := lex.TakeParameter()
			if err != nil {
				t.Fatalf("Error on calling Tokenizer.TakeParameter: %s\n", err.Error())
			}
			if param != "input" {
				t.Fatalf("'%s' does not match the expected parameter 'input'.\n", param)
			}
		}
	}
loopExit:

	if !optHelp { t.Error("Expected help option not encountered.\n") }
	if !optVersion { t.Error("Expected version option not encountered.\n") }
	if !optX { t.Error("Expected x option not encountered.\n") }
	if !optZ { t.Error("Expected z option not encountered.\n") }
	if !optV { t.Error("Expected v option not encountered.\n") }
	if !optF { t.Error("Expected f option not encountered.\n") }
	if !optOutput { t.Error("Expected output option not encountered.\n") }
	if !optInput { t.Error("Expected input option not encountered.\n") }
}

