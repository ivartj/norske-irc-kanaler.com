package bbgo

import (
	"strings"
	"testing"
	"os"
)

func TestGetOpenTagNameAndArg(t *testing.T) {
	t1 := "[name]"
	name, arg := getOpenTagNameAndArg([]rune(t1))
	if name != "name" {
		t.Errorf("\"name\" != \"%s\"\n", name)
	}

	if arg != "" {
		t.Errorf("\"arg\" != \"%s\"\n", arg)
	}

	t2 := "[name=arg]"
	name, arg = getOpenTagNameAndArg([]rune(t2))
	if name != "name" {
		t.Errorf("\"name\" != \"%s\"\n", name)
	}

	if arg != "arg" {
		t.Errorf("\"arg\" != \"%s\"\n", arg)
	}
}

func TestParser(t *testing.T) {
	str := `[i]test[quote]best[/quote]`
	input := strings.NewReader(str)

	ch := lex(input)
	p := newParser(ch, os.Stdout)
	err := p.parse()
	if err != nil {
		t.Error(err)
	}
}

