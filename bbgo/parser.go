package bbgo

import (
	"errors"
	"fmt"
	"html"
	"io"
	"strings"
)

type parser struct {
	input     <-chan token
	output    io.Writer
	blockRoll []*blockScope
	newline   bool
	isInline  bool
	blockLine *blockLineTag
}

type blockScope struct {
	tag        *blockTag
	inlineRoll []*inlineTag
	blockLine  *blockLineTag
}

func newParser(input <-chan token, output io.Writer) *parser {
	return &parser{
		input:  input,
		output: output,
		blockRoll: []*blockScope{
			&blockScope{
				inlineRoll: []*inlineTag{},
			},
		},
	}
}

func (p *parser) scope() *blockScope {
	return p.blockRoll[len(p.blockRoll)-1]
}

func (p *parser) parse() (errval error) {
	defer func() {
		err, ok := recover().(error)
		if ok {
			errval = err
		}
	}()

	for {
		t, eof := p.next()
		if eof {
			p.closeBlockLine()
			for len(p.blockRoll) > 1 {
				p.closeBlock()
			}
			return nil
		}

		switch t.typ {
		case textToken:
			p.openParagraph()
			fmt.Fprintf(p.output, "%s", html.EscapeString(string(t.raw)))
		case openTagToken:
			name, arg := getOpenTagNameAndArg(t.raw)
			p.processOpenTag(name, arg)
		case closeTagToken:
			name := getCloseTagName(t.raw)
			p.processCloseTag(name)
		case newlineToken:
			if p.newline {
				p.closeBlockLine()
			}
			p.newline = !p.newline
		}
	}

	return nil
}

func (p *parser) openParagraph() {
	typ := &simpleBlockLineTagType{"p"}
	p.openBlockLine(newBlockLineTag(typ, "p", ""))
}

func (p *parser) openBlockLine(tag *blockLineTag) {
	if p.isInline {
		if p.newline {
			fmt.Fprintln(p.output, "<br />")
			p.newline = false
		}
		return
	}

	tag.printOpen(p.output)

	for _, scope := range p.blockRoll {
		for _, inlineTag := range scope.inlineRoll {
			inlineTag.printOpen(p.output)
		}
	}

	p.newline = false
	p.isInline = true
	p.blockLine = tag
}

func (p *parser) closeBlockLine() {
	if !p.isInline {
		return
	}

	for i := len(p.blockRoll) - 1; i >= 0; i-- {
		scope := p.blockRoll[i]
		for j := len(scope.inlineRoll) - 1; j >= 0; j-- {
			inlineTag := scope.inlineRoll[j]
			inlineTag.printClose(p.output)
		}
	}

	p.newline = false
	p.isInline = false
	p.blockLine.printClose(p.output)
}

func (p *parser) openBlock(tag *blockTag) {
	p.closeBlockLine()
	p.blockRoll = append(p.blockRoll, &blockScope{
		tag:        tag,
		inlineRoll: []*inlineTag{},
	})
	tag.printOpen(p.output)
}

func (p *parser) closeBlock() {
	p.closeBlockLine()
	if len(p.blockRoll) <= 1 {
		panic("Closing root scope")
	}
	p.blockRoll[len(p.blockRoll)-1].tag.printClose(p.output)
	p.blockRoll = p.blockRoll[:len(p.blockRoll)-1]
}

func (p *parser) openInline(tag *inlineTag) {
	p.openParagraph()
	p.scope().inlineRoll = append(p.scope().inlineRoll, tag)
	tag.printOpen(p.output)
}

func (p *parser) closeInline(name string) {
	// TODO: currently only considers previous tag on same scope

	p.openParagraph()

	rollLen := len(p.scope().inlineRoll)
	if rollLen == 0 {
		panic("Empty inline roll")
	}

	tag := p.scope().inlineRoll[rollLen-1]
	if tag.name != name {
		panic(fmt.Errorf("\"%s\" != \"%s\"", tag.name, name))
	}

	tag.printClose(p.output)

	p.scope().inlineRoll = p.scope().inlineRoll[:rollLen-1]
}

func (p *parser) processOpenTag(name, arg string) {
	blockType, isBlock := blockTagTypes[name]
	inlineType, isInline := inlineTagTypes[name]
	blockLineType, isBlockLine := blockLineTagTypes[name]
	blockContentType, isBlockContent := blockContentTagTypes[name]
	inlineContentType, isInlineContent := inlineContentTagTypes[name]

	switch {
	case isBlock:
		tag := newBlockTag(blockType, name, arg)
		p.openBlock(tag)
	case isInline:
		tag := newInlineTag(inlineType, name, arg)
		p.openInline(tag)
	case isBlockLine:
		tag := newBlockLineTag(blockLineType, name, arg)
		p.openBlockLine(tag)
	case isBlockContent:
		fmt.Println(blockContentType)
	case isInlineContent:
		fmt.Println(inlineContentType)
	}
}

func (p *parser) processCloseTag(name string) {
	_, isBlock := blockTagTypes[name]
	_, isInline := inlineTagTypes[name]
	_, isBlockLine := blockLineTagTypes[name]
	blockContentType, isBlockContent := blockContentTagTypes[name]
	inlineContentType, isInlineContent := inlineContentTagTypes[name]

	switch {
	case isBlock:
		// TODO: Check that the name matches
		p.closeBlock()
	case isInline:
		p.closeInline(name)
	case isBlockLine:
		p.closeBlockLine()
	case isBlockContent:
		fmt.Println(blockContentType)
	case isInlineContent:
		fmt.Println(inlineContentType)
	}
}

func getCloseTagName(tag []rune) string {
	return string(tag[2 : len(tag)-1])
}

func getOpenTagNameAndArg(tag []rune) (name string, arg string) {
	var i int
	var r rune

	for i, r = range tag[1:] {
		if r == ']' || r == '=' {
			name = strings.ToLower(string(tag[1 : i+1]))
			break
		}
	}
	if r == ']' {
		return name, ""
	}
	if r != '=' {
		panic(fmt.Errorf("Invalid tag token, '%s'", string(tag)))
	}

	escape := false
	escapeNext := false
	argrunes := []rune{}
	for _, r = range tag[i+2:] {
		escape = escapeNext
		escapeNext = false
		switch r {
		case '\\':
			if escape {
				argrunes = append(argrunes, '\\')
			} else {
				escapeNext = true
			}
		case ']':
			if escape {
				argrunes = append(argrunes, ']')
			} else {
				return name, string(argrunes)
			}
		default:
			argrunes = append(argrunes, r)
		}
		escape = escapeNext
	}

	panic(fmt.Errorf("Invalid tag token, '%s'", string(tag)))
	return "", ""
}

func (p *parser) next() (token, bool) {
	t, notEOF := <-p.input
	if notEOF == false {
		return t, true
	}

	if t.typ == errorToken {
		panic(errors.New(string(t.raw)))
	}

	return t, false
}
