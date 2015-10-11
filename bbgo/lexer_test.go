package bbgo

import (
	"testing"
	"strings"
	"fmt"
)

func TestLexer(t *testing.T) {
	const input string = `
[url=http://www.google.com/]www.google.com[/url]

[quote=Abraham Lincoln]Don't believe every quote your read on the Internet.[/quote]
`
	output := lex(strings.NewReader(input))
	for v := range output {
		fmt.Printf("%d %q\n", v.typ, v.raw)
	}
}
