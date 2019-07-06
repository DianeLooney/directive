package ast

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

var logitI = 0

func logit() func() {
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	path := f.Name()
	split := strings.Split(path, "/")
	s := split[len(split)-1]

	fmt.Printf("%s+%s\n", strings.Repeat("\t", logitI), s)
	logitI++
	return func() {
		logitI--
		fmt.Printf("%s-%s\n", strings.Repeat("\t", logitI), s)
	}
}

type Token uint8

const (
	STRING_DBL = iota
	STRING_SNG
	STRING_LIT
	IDENT
	AT
	LCURLY
	RCURLY
	LSQUARE
	RSQUARE
	SEMI
	COMMA
)

type Position struct {
	Byte   int
	Line   int
	Column int
}

type Node interface {
	Begin() Position
	End() Position
	Text() string
}

type node struct {
	begin Position
	end   Position
	text  []byte
}

func (n node) Begin() Position {
	return n.begin
}

func (n node) End() Position {
	return n.end
}

func (n node) Text() string {
	return string(n.text)
}

type Document struct {
	node
	Directives []Node
}

func (d Document) String() string {
	strs := make([]string, len(d.Directives))
	for i, r := range d.Directives {
		strs[i] = fmt.Sprintf("%s", r)
	}
	return strings.Join(strs, " ")
}

type Object struct {
	node
	Directives []Node
}

func (o Object) String() string {
	strs := make([]string, len(o.Directives))
	for i, r := range o.Directives {
		strs[i] = fmt.Sprintf("%s", r)
	}
	return fmt.Sprintf("{%s}", strings.Join(strs, " "))
}

type String struct {
	node
	Value string
}

func (s String) String() string {
	return s.Value
}

type Directive struct {
	node
	Identifier string
	IsContext  bool
	Value      Node
}

func (d Directive) String() string {
	if d.IsContext {
		return fmt.Sprintf("@%s %s;", d.Identifier, d.Value)
	} else {
		return fmt.Sprintf("%s %s;", d.Identifier, d.Value)
	}
}

type RepeatedDirective struct {
	node
	Identifier string
	IsContext  bool
	Values     []Node
}

func (d RepeatedDirective) String() string {
	strs := make([]string, len(d.Values))
	for i, v := range d.Values {
		strs[i] = fmt.Sprintf("%s", v)
	}
	s := strings.Join(strs, ", ")
	if d.IsContext {
		return fmt.Sprintf("[@%s %s]; ", d.Identifier, s)
	} else {
		return fmt.Sprintf("[%s %s]; ", d.Identifier, s)
	}
}

type Parser struct {
	data []byte
}

func NewParser(data []byte) *Parser {
	return &Parser{
		data: data,
	}
}

func (p *Parser) Parse() (Node, error) {
	return p.parseDocument()
}

var whitespace = map[byte]bool{
	' ':  true,
	'\t': true,
	'\r': true,
	'\n': true,
}

func (p *Parser) peekByte() (byte, bool) {
	if len(p.data) == 0 {
		return ' ', false
	}

	return p.data[0], true
}

func (p *Parser) consumeByte(b byte) error {
	if p.data[0] != b {
		return fmt.Errorf("expected byte '%v'", b)
	}

	p.data = p.data[1:]

	return nil
}

func (p *Parser) consumeRegex(r *regexp.Regexp) (text string, err error) {
	s := r.FindSubmatch(p.data)
	if s == nil {
		return "", fmt.Errorf("expected to match regex '%s'", r)
	}

	text = string(s[0])
	p.data = p.data[len(text):]

	return
}

func (p *Parser) skipWhitespace() {
	for {
		c, ok := p.peekByte()
		if ok && whitespace[c] {
			p.consumeByte(c)
			continue
		}
		break
	}
	return
}

func (p *Parser) skipSemi() {
	if b, _ := p.peekByte(); b == ';' {
		p.consumeByte(';')
	}
}

func (p *Parser) skipComma() {
	if b, _ := p.peekByte(); b == ',' {
		p.consumeByte(';')
	}
}

func (p *Parser) parseDocument() (d *Document, err error) {
	defer logit()()

	d = &Document{}
	for {
		p.skipWhitespace()

		c, ok := p.peekByte()
		if !ok {
			return
		}

		if c == '[' {
			v, err := p.parseRepeatedDirective()
			if err != nil {
				return d, err
			}
			d.Directives = append(d.Directives, v)
		} else {
			v, err := p.parseDirective()
			if err != nil {
				return d, err
			}
			d.Directives = append(d.Directives, v)
		}
	}
}

func (p *Parser) parseDirective() (d *Directive, err error) {
	defer logit()()

	d = &Directive{}
	c, ok := p.peekByte()
	if !ok {
		return nil, fmt.Errorf("encountered EOF while parsing Directive")
	}
	if c == '@' {
		d.IsContext = true
		d.Identifier, err = p.parseContext()
		if err != nil {
			return nil, err
		}
	} else {
		d.Identifier, err = p.parseIdentifier()
		if err != nil {
			return nil, err
		}
	}

	p.skipWhitespace()

	d.Value, err = p.parseValue()

	p.skipSemi()

	return d, err
}

func (p *Parser) parseRepeatedDirective() (d *RepeatedDirective, err error) {
	err = p.consumeByte('[')
	if err != nil {
		return nil, fmt.Errorf("could not parse RepeatedDirective: %v", err)
	}
	d = &RepeatedDirective{}

	if c, ok := p.peekByte(); !ok {
		return nil, fmt.Errorf("encountered EOF while parsing RepeatedDirective")
	} else if c == '@' {
		d.Identifier, err = p.parseContext()
	} else {
		d.Identifier, err = p.parseIdentifier()
	}

	if err != nil {
		return nil, fmt.Errorf("could not parse RepeatedDirective.Identifier: %v", err)
	}

	for {
		p.skipWhitespace()

		if c, ok := p.peekByte(); !ok {
			return nil, fmt.Errorf("encountered EOF while parsing RepeatedDirective")
		} else if c == ']' {
			p.consumeByte(']')
			break
		}

		v, err := p.parseValue()
		if err != nil {
			return d, err
		}
		d.Values = append(d.Values, v)

		if c, ok := p.peekByte(); !ok {
			return nil, fmt.Errorf("encountered EOF while parsing RepeatedDirective")
		} else if c == ']' {
			p.consumeByte(']')
			break
		}
	}
	p.skipSemi()
	return d, nil
}

var identifier = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]+`)

func (p *Parser) parseIdentifier() (ident string, err error) {
	defer logit()()

	ident, err = p.consumeRegex(identifier)
	if err != nil {
		err = fmt.Errorf("Unable to parse identifier: %v, but it was '%s'", err, p.data[:10])
	}
	return
}

func (p *Parser) parseContext() (ident string, err error) {
	err = p.consumeByte('@')
	if err != nil {
		return "", fmt.Errorf("could not parse Context: %v", err)
	}
	return p.consumeRegex(identifier)
}

var strDbl = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
var strSgl = regexp.MustCompile(`'((?:[^"\\]|\\.)*)'`)
var strLit = regexp.MustCompile("`" + `([^` + "`" + `]*)` + "`")

func (p *Parser) parseValue() (v Node, err error) {
	defer logit()()
	c, ok := p.peekByte()

	if !ok {
		return nil, fmt.Errorf("encountered EOF while parsing Value")
	} else if c == '"' {
		s, err := p.consumeRegex(strDbl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Value: %v", err)
		}
		return &String{Value: s}, nil
	} else if c == '{' {
		o, err := p.parseObject()
		if err != nil {
			return nil, fmt.Errorf("failed to parse Value: %v", err)
		}
		return o, nil
	}
	return nil, fmt.Errorf("failed to parse Value: unrecognized character %s", []byte{c})
}

func (p *Parser) parseString() (s *String, err error) {
	defer logit()()

	c, ok := p.peekByte()
	if !ok {
		return nil, fmt.Errorf("encountered EOF while parsing Value")
	}

	if c == '"' {
		s, err := p.consumeRegex(strDbl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Value: %v", err)
		}
		return &String{Value: s}, nil
	}

	if c == '\'' {
		s, err := p.consumeRegex(strSgl)
		if err != nil {
			return nil, err
		}
		return &String{Value: s}, nil
	}

	if c == '`' {
		s, err := p.consumeRegex(strLit)
		if err != nil {
			return nil, err
		}
		return &String{Value: s}, nil
	}

	p.consumeByte(c)
	return nil, fmt.Errorf("encountered unexpected byte %v when parsing for a string", c)
}

func (p *Parser) parseObject() (o *Object, err error) {
	p.consumeByte('{')
	o = &Object{}
	for {
		p.skipWhitespace()

		if c, ok := p.peekByte(); !ok {
			return nil, fmt.Errorf("encountered EOF while parsing Object")
		} else if c == '}' {
			p.consumeByte('}')
			return o, nil
		} else if c == '[' {
			v, err := p.parseRepeatedDirective()
			if err != nil {
				return o, err
			}
			o.Directives = append(o.Directives, v)
		} else {
			v, err := p.parseDirective()
			if err != nil {
				return o, err
			}
			o.Directives = append(o.Directives, v)
		}
	}
}