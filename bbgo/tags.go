package bbgo

import (
	"io"
	"fmt"
	"html"
)

const (
	BLOCK_TAG int = iota
	INLINE_TAG
	BLOCK_LINE_TAG
	BLOCK_CONTENT_TAG
	INLINE_CONTENT_TAG
)

var inlineTagTypes = map[string]InlineTagType{
	"i" : &italicsTagType{},
	"b" : &boldTagType{},
	"url" : &urlTagType{},
}

var blockTagTypes = map[string]BlockTagType{
	"quote" : &quoteTagType{},
}

var blockLineTagTypes = map[string]BlockLineTagType{
	"h1" : &simpleBlockLineTagType{"h1"},
	"h2" : &simpleBlockLineTagType{"h2"},
	"h3" : &simpleBlockLineTagType{"h3"},
}

var inlineContentTagTypes = map[string]InlineContentTagType{
	"img" : &imgTagType{},
}

var blockContentTagTypes = map[string]BlockContentTagType{
}

type InlineTagType interface {
	PrintOpen(w io.Writer, arg string)
	PrintClose(w io.Writer, arg string)
}

type BlockTagType interface {
	PrintOpen(w io.Writer, arg string)
	PrintClose(w io.Writer, arg string)
}

type BlockLineTagType interface {
	PrintOpen(w io.Writer, arg string)
	PrintClose(w io.Writer, arg string)
}

type InlineContentTagType interface {
	Print(w io.Writer, arg, content string) 
}

type BlockContentTagType interface {
	Print(w io.Writer, arg, content string) 
}

func RegisterInlineTagType(name string, tagType InlineTagType) {
	inlineTagTypes[name] = tagType
}

func RegisterBlockTagType(name string, tagType BlockTagType) {
	blockTagTypes[name] = tagType
}

func RegisterInlineContentTagType(name string, tagType InlineContentTagType) {
	inlineContentTagTypes[name] = tagType
}

func RegisterBlockContentTagType(name string, tagType BlockContentTagType) {
	blockContentTagTypes[name] = tagType
}

type imgTagType struct{}

func (img *imgTagType) Print(w io.Writer, arg, content string) {
	// TODO: Further sanitation
	fmt.Fprintf(w, "<img src=\"%s\" alt=\"\" />", html.EscapeString(content))
} 

type quoteTagType struct{}

func (quote *quoteTagType) PrintOpen(w io.Writer, arg string) {
	if arg != "" {
		fmt.Fprintf(w, "<p>%s wrote:</p>\n", html.EscapeString(arg))
	}
	fmt.Fprintln(w, "<blockquote>")
}

func (quote *quoteTagType) PrintClose(w io.Writer, arg string) {
	fmt.Fprintln(w, "</blockquote>")
}

type italicsTagType struct{}

func (italics *italicsTagType) PrintOpen(w io.Writer, arg string) {
	fmt.Fprint(w, "<em>")
}

func (italics *italicsTagType) PrintClose(w io.Writer, arg string) {
	fmt.Fprint(w, "</em>")
}

type boldTagType struct{}

func (bold *boldTagType) PrintOpen(w io.Writer, arg string) {
	fmt.Fprint(w, "<strong>")
}

func (bold *boldTagType) PrintClose(w io.Writer, arg string) {
	fmt.Fprint(w, "</strong>")
}

type urlTagType struct{}

func (url *urlTagType) PrintOpen(w io.Writer, arg string) {
	fmt.Fprintf(w, "<a href=\"%s\">", html.EscapeString(arg))
}

func (url *urlTagType) PrintClose(w io.Writer, arg string) {
	fmt.Fprint(w, "</a>")
}

type blockTag struct {
	typ BlockTagType
	name, arg string
}

func newBlockTag(typ BlockTagType, name, arg string) *blockTag {
	return &blockTag{typ, name, arg}
}

func (tag *blockTag) printOpen(w io.Writer) {
	tag.typ.PrintOpen(w, tag.arg)
}

func (tag *blockTag) printClose(w io.Writer) {
	tag.typ.PrintClose(w, tag.arg)
}

type inlineTag struct {
	typ InlineTagType
	name, arg string
}

func newInlineTag(typ InlineTagType, name, arg string) *inlineTag {
	return &inlineTag{typ, name, arg}
}

func (tag *inlineTag) printOpen(w io.Writer) {
	tag.typ.PrintOpen(w, tag.arg)
}

func (tag *inlineTag) printClose(w io.Writer) {
	tag.typ.PrintClose(w, tag.arg)

}

type blockLineTag struct {
	typ BlockLineTagType
	name, arg string
}

func newBlockLineTag(typ BlockLineTagType, name, arg string) *blockLineTag {
	return &blockLineTag{typ, name, arg}
}

func (tag *blockLineTag) printOpen(w io.Writer) {
	tag.typ.PrintOpen(w, tag.arg)
}

func (tag *blockLineTag) printClose(w io.Writer) {
	tag.typ.PrintClose(w, tag.arg)
}

type simpleBlockLineTagType struct {
	name string
}

func (tag *simpleBlockLineTagType) PrintOpen(w io.Writer, arg string) {
	fmt.Fprintf(w, "<%s>", tag.name)
}

func (tag *simpleBlockLineTagType) PrintClose(w io.Writer, arg string) {
	fmt.Fprintf(w, "</%s>", tag.name)
}

