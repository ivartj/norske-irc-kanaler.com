package args

import (
	"fmt"
	"errors"
	"io"
)

const (
	End			= -1 - iota
	Plain			= -1 - iota
	Unrecognized		= -1 - iota
	MissingOptionArgument	= -1 - iota
)

type option struct {
	Code int
	Short int
	Long string
	Argument string
	Help string
}

type Parser struct {
	options []option
	args []string
	offset int
	argOffset int
	noMoreOptions bool
}

func NewParser(args []string) *Parser {
	return &Parser{
		args: args,
	}
}

func (p *Parser) AddOption(code int, short int, long string, argument string, help string) {
	p.options = append(p.options, option{code, short, long, argument, help})
}

func (p *Parser) SetArgs(args []string) {
	p.args = args
	p.offset = 0
	p.argOffset = 0
}

func (p *Parser) getLongOption(long string) (*option, error) {
	for _, v := range p.options {
		if v.Long == long {
			return &v, nil
		}
	}

	return nil, errors.New("Not found")
}

func (p *Parser) getShortOption(short int) (*option, error) {
	for _, v := range p.options {
		if v.Short == short {
			return &v, nil
		}
	}

	return nil, errors.New("Not found")
}

func (p *Parser) Parse() (int, string) {

	// no more arguments
	if p.offset == len(p.args) {
		return End, ""
	}
	arg := p.args[p.offset]

	// plain argument
	if p.noMoreOptions || len(arg) <= 1 || arg[0] != '-' {
		p.offset++
		return Plain, arg
	}

	// options terminator
	if arg == "--" {
		p.offset++
		p.noMoreOptions = true
		return p.Parse() // recurse once
	}

	// long option
	// TODO: implement --config=arg style option arguments
	if arg[:2] == "--" {
		p.offset++
		opt, err := p.getLongOption(arg[2:])
		if err != nil {
			return Unrecognized, arg
		}

		if opt.Argument == "" {
			return opt.Code, ""
		} else {
			if p.offset == len(p.args) {
				return MissingOptionArgument, arg
			}
			p.offset++
			return opt.Code, p.args[p.offset - 1]
		}
	}

	// short option
	p.argOffset++
	if p.argOffset + 1 == len(arg) {
		defer func() { p.argOffset = 0 }()
		p.offset++
	}

	opt, err := p.getShortOption(int(arg[p.argOffset]))
	if err != nil {
		return Unrecognized, fmt.Sprintf("-%c", arg[p.argOffset])
	}

	if opt.Argument == "" {
		return opt.Code, ""
	} else {
		if p.offset == len(p.args) || p.argOffset + 1 != len(arg) {
			return MissingOptionArgument, fmt.Sprintf("-%c", arg[p.argOffset])
		}
		p.offset++
		return opt.Code, p.args[p.offset - 1]
	}
}

func optionSynopsis(o option) string {

	short := o.Short != '-'
	long := o.Long != ""
	arg := o.Argument != ""

	switch {
	case short && !long && !arg:
		return fmt.Sprintf("  -%c  ", o.Short)
	case short && long && !arg:
		return fmt.Sprintf("  -%c, --%s  ", o.Short, o.Long)
	case short && long && arg:
		return fmt.Sprintf("  -%c, --%s=%s  ", o.Short, o.Long, o.Argument)
	case short && !long && arg:
		return fmt.Sprintf("  -%c %s  ", o.Short, o.Argument)
	case !short && long && !arg:
		return fmt.Sprintf("  --%s  ", o.Long)
	case !short && long && arg:
		return fmt.Sprintf("  --%s=%s  ", o.Long, o.Argument)
	}

	// TODO: place this panic in AddOption instead with an explanation.
	panic("Not a valid option type.")
}

func (p *Parser) PrintUsage(w io.Writer) {

	optHelpOffset := 0

	for _, v := range p.options {

		off := len(optionSynopsis(v))

		if optHelpOffset < off {
			optHelpOffset = off	
		}
	}

	for _, v := range p.options {

		off, _ := fmt.Fprint(w, optionSynopsis(v))

		for i := off; i < optHelpOffset; i++ {
			fmt.Fprintf(w, " ")
		}

		fmt.Fprintf(w, "%s\n", v.Help)
	}
}

type Tokenizer struct {
	
}

func NewTokenizer(argv []string) *Tokenizer {
	return &Tokenizer{}
}

func (tok *Tokenizer) Next() (string, error) {
	return "", errors.New("Unimplemented")
}

func (tok *Tokenizer) TakeParameter() (string, error) {
	return "", errors.New("Unimplemented")
}

