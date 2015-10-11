package bbgo

import (
	"io"
)

func Process(src io.Reader, dest io.Writer) error {
	ch := lex(src)
	p := newParser(ch, dest)
	return p.parse()
}

