package bbgo

import (
	"io"
	"bufio"
	"runtime"
	"fmt"
	"unicode/utf8"
	"unicode"
)

type lexer struct {
	pos int
	lexeme []rune
	input *bufio.Reader
	output chan <-token
	atEOF bool
}

func lex(r io.Reader) <-chan token {
	ch := make(chan token, 100)
	l := &lexer{
		input: bufio.NewReader(r),
		output: ch,
		lexeme: []rune{},
	}
	go l.run()
	return ch
}

func (l *lexer) run() {
	fn := l.stateStart

	// Stops when lexer.stop() is called
	for {
		fn = fn()
	}
}

func (l *lexer) stop() {
	close(l.output)
	runtime.Goexit()
}

func (l *lexer) fail(format string, args... interface{}) {
	l.output <- token{
		typ: errorToken,
		raw: []rune(fmt.Sprintf(format, args...)),
	}
	close(l.output)
	runtime.Goexit()
}

func (l *lexer) next() (rune, bool) {
	if l.pos < len(l.lexeme) {
		l.pos++
		return l.lexeme[l.pos - 1], false
	}

	r, _, err := l.input.ReadRune()
	if err == io.EOF {
		return 0, true
	} else if err != nil {
		l.fail("%s", err.Error())
	}

	l.lexeme = append(l.lexeme, r)
	l.pos++

	return r, false
}

func (l *lexer) back() {
	l.pos--
}

func (l *lexer) peek() (rune, bool) {
	if l.pos < len(l.lexeme) {
		return l.lexeme[l.pos], false
	}
	bytes, err := l.input.Peek(4)
	if err == io.EOF {
		return 0, true
	}
	if err != nil {
		l.fail("%s", err.Error())
	}
	r, _ := utf8.DecodeRune(bytes)
	return r, false
}

func (l *lexer) emit(typ tokenType) {
	l.output <- token{ typ, l.lexeme[:l.pos] }
	l.lexeme = l.lexeme[l.pos:]
	l.pos = 0
}


type lexerStateFn func() lexerStateFn

func (l *lexer) stateStart() lexerStateFn {
	for {
		r, eof := l.next()
		if eof {
			l.stop()
		}

		switch r {
		case '[':
			return l.stateAfterSquareBracket
		case '\n':
			l.emit(newlineToken)
		case '\r':
			break
		default:
			return l.stateText
		}
	}
}

func (l *lexer) stateAfterSquareBracket() lexerStateFn {
	for {
		r, eof := l.next()
		if eof {
			return l.stateText
		}

		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			return l.stateOpeningTagName
		case r == '/':
			return l.stateClosingTagName
		default:
			return l.stateText
		}
	}
}

func (l *lexer) stateClosingTagName() lexerStateFn {
	for {
		r, eof := l.next()
		if eof {
			return l.stateText
		}

		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			continue
		case r == ']':
			l.emit(closeTagToken)
			return l.stateStart
		default:
			return l.stateText
		}
	}
}

func (l *lexer) stateOpeningTagName() lexerStateFn {
	for {
		r, eof := l.next()
		if eof {
			return l.stateText
		}

		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			continue
		case r == ']':
			l.emit(openTagToken)
			return l.stateStart
		case r == '=':
			return l.stateSingleTagArgument
		default:
			return l.stateText
		}
	}
}

func (l *lexer) stateSingleTagArgument() lexerStateFn {

	escape := false
	escapeNext := false

	for {
		r, eof := l.next()
		if eof {
			return l.stateText
		}

		escapeNext = false

		switch {
		case r == '\\':
			if !escape { escapeNext = true }
		case r == '[':
			if escape { continue }
			return l.stateText
		case r == ']':
			if escape { continue }
			l.emit(openTagToken)
			return l.stateStart
		case unicode.IsPrint(r):
			continue
		default:
			return l.stateText
		}

		escape = escapeNext
	}
}

func (l *lexer) stateText() lexerStateFn {
	for {
		r, eof := l.next()
		if eof {
			l.emit(textToken)
			l.stop()
		}

		switch r {
		case '[':
			l.back()
			l.emit(textToken)
			return l.stateStart
		case '\n':
			l.back()
			l.emit(textToken)
			return l.stateStart
		default:
			
		}
	}
}

type tokenType int

type token struct {
	typ tokenType
	raw []rune
}

const (
	errorToken tokenType 	= iota
	openTagToken
	closeTagToken
	textToken
	newlineToken
)

