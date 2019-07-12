package ast

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

var logitI = 0

const loggingEnabled = false

func logit() func() {
	if !loggingEnabled {
		return func() {}
	}

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
	Execute(x interface{}) error
}

type Whitespace struct {
	node
}

func (w Whitespace) Execute(x interface{}) error {
	return nil
}

type node struct {
	begin Position
	end   Position
	text  []byte
}

func (n node) Execute(x interface{}) error {
	panic("execute is not defined yet")
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

func (d Document) Execute(x interface{}) error {
	for _, dir := range d.Directives {
		err := dir.Execute(x)
		if err != nil {
			return err
		}
	}
	return nil
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

type Number struct {
	node
	Value string
}

type Note struct {
	node
	Value string
}

type Unknown struct {
	node
	Value string
}

func (n Number) String() string {
	return n.Value
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

func (d Directive) Execute(x interface{}) error {
	switch v := d.Value.(type) {
	case *Object:
		y, err := get(x, d.Identifier)
		if err != nil {
			return err
		}
		return v.Execute(y)
	case *String:
		return set(x, d.Identifier, v.Value)
	case *Number:
		return set(x, d.Identifier, v.Value)
	case *Note:
		return set(x, d.Identifier, v.Value)
	case *Unknown:
		return set(x, d.Identifier, v.Value)
	}
	log.Fatalf("Unhandled value type %T", d.Value)
	panic("unreachable")
}

func (o Object) Execute(x interface{}) error {
	for _, dir := range o.Directives {
		err := dir.Execute(x)
		if err != nil {
			return err
		}
	}
	return nil
}

type RepeatedDirective struct {
	node
	Identifier string
	IsContext  bool
	Values     []Node
}

func (r RepeatedDirective) Execute(x interface{}) (err error) {
	defer func() {
		e := recover()
		if err == nil && e != nil {
			err = fmt.Errorf("Recovered from panic: %v", e)
		}
	}()

	for _, value := range r.Values {
		switch v := value.(type) {
		case *Object:
			y, err := get(x, r.Identifier)
			if err != nil {
				return err
			}
			v.Execute(y)
			continue
		case *String:
			err := set(x, r.Identifier, v.Value)
			if err != nil {
				return err
			}
			continue
		case *Number:
			err := set(x, r.Identifier, v.Value)
			if err != nil {
				return err
			}
			continue
		case *Note:
			err := set(x, r.Identifier, v.Value)
			if err != nil {
				return err
			}
			continue
		case *Unknown:
			err := set(x, r.Identifier, v.Value)
			if err != nil {
				return err
			}
			continue
		default:
			log.Fatalf("Unhandled value type %T", v)
		}
		panic("unreachable")
	}
	return nil
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

func (p *Parser) skipWhitespace() (n []Node) {
	firstNewline := true
	for {
		c, ok := p.peekByte()
		if c == '\n' {
			if !firstNewline {
				n = append(n, Whitespace{})
			}
			firstNewline = false
		}
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
		d.Directives = append(d.Directives, p.skipWhitespace()...)

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
	} else if c == '"' || c == '\'' || c == '`' {
		s, err := p.parseString()
		if err != nil {
			return nil, fmt.Errorf("failed to parse Value: %v", err)
		}
		return s, nil
	} else if c == '{' {
		o, err := p.parseObject()
		if err != nil {
			return nil, fmt.Errorf("failed to parse Value: %v", err)
		}
		return o, nil
	} else if unicode.IsNumber(rune(c)) || c == '-' || c == '+' {
		n, err := p.parseNumber()
		if err != nil {
			n, err := p.parseNote()
			if err != nil {
				return nil, fmt.Errorf("failed to parse Note: %v", err)
			}
			return n, nil
		}
		return n, nil
	} else if c == '?' {
		n, err := p.parseUnknown()
		if err != nil {
			return nil, fmt.Errorf("failed to parse Unknown: %v", err)
		}
		return n, nil
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
		v, err := strconv.Unquote(s)
		if err != nil {
			return nil, err
		}
		return &String{Value: v, node: node{text: []byte(s)}}, nil
	}

	if c == '\'' {
		s, err := p.consumeRegex(strSgl)
		if err != nil {
			return nil, err
		}
		v, err := strconv.Unquote(s)
		if err != nil {
			return nil, err
		}
		return &String{Value: v, node: node{text: []byte(s)}}, nil
	}

	if c == '`' {
		s, err := p.consumeRegex(strLit)
		if err != nil {
			return nil, err
		}
		v, err := strconv.Unquote(s)
		if err != nil {
			return nil, err
		}
		return &String{Value: v, node: node{text: []byte(s)}}, nil
	}

	p.consumeByte(c)
	return nil, fmt.Errorf("encountered unexpected byte %v when parsing for a string", c)
}

func (p *Parser) parseObject() (o *Object, err error) {
	p.consumeByte('{')
	o = &Object{}
	for {
		o.Directives = append(o.Directives, p.skipWhitespace()...)

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

var number = regexp.MustCompile(`^([+-]?)[0-9]+(?:\.[0-9]*)\b?`)

func (p *Parser) parseNumber() (n *Number, err error) {
	defer logit()()

	num, err := p.consumeRegex(number)
	if err != nil {
		return nil, fmt.Errorf("Number was not formatted correctly: %v", err)
	}
	n = &Number{
		Value: num,
		node:  node{text: []byte(num)},
	}

	return n, nil
}

var note = regexp.MustCompile(`^([+-]?[0-9]+[#b]*)\b`)

func (p *Parser) parseNote() (n *Note, err error) {
	defer logit()()

	v, err := p.consumeRegex(note)
	if err != nil {
		return nil, fmt.Errorf("Note was not formatted correctly: %v", err)
	}
	n = &Note{
		Value: v,
		node:  node{text: []byte(v)},
	}

	return n, nil
}

var unknown = regexp.MustCompile(`^(\?)`)

func (p *Parser) parseUnknown() (n *Unknown, err error) {
	defer logit()()

	v, err := p.consumeRegex(unknown)
	if err != nil {
		return nil, fmt.Errorf("Unknown was not formatted correctly: %v", err)
	}
	n = &Unknown{
		Value: v,
		node:  node{text: []byte(v)},
	}

	return n, nil
}

func get(x interface{}, field string) (interface{}, error) {
	t := reflect.ValueOf(x)
	m := t.MethodByName(field)
	if !m.IsNil() {
		out := m.Call(nil)
		return out[0].Interface(), nil
	}

	f := t.FieldByName(field)
	if !f.IsNil() {
		return f.Interface(), nil
	}

	return nil, fmt.Errorf("%T did not have method or field %s", x, field)
}

func set(x interface{}, field string, value string) error {
	t := reflect.ValueOf(x)

	m := t.MethodByName(field)
	if k := m.Kind(); k == reflect.Func {
		t := m.Type().In(0)
		switch t.Kind() {
		case reflect.Float64:
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("error while calling method %s with '%s': %v", field, value, err)
			}
			m.Call([]reflect.Value{reflect.ValueOf(v)})
		case reflect.String:
			m.Call([]reflect.Value{reflect.ValueOf(value)})
		case reflect.Int:
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return fmt.Errorf("error while calling method %s with '%s': %v", field, value, err)
			}
			m.Call([]reflect.Value{reflect.ValueOf(int(v))})
		default:
			log.Fatalf("Need to implement kind '%s' in directive/ast.set - method", t.Kind())
		}
		return nil
	}

	f := t.Elem().FieldByName(field)
	zero := reflect.Value{}
	if f != zero {
		switch f.Kind() {
		case reflect.String:
			f.Set(reflect.ValueOf(value))
		case reflect.Int:
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("error while setting field %s to '%s': %v", field, value, err)
			}
			f.Set(reflect.ValueOf(v))
		case reflect.Float64:
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("error while setting field %s to '%s': %v", field, value, err)
			}
			f.Set(reflect.ValueOf(v))
		default:
			log.Fatalf("Need to implement kind '%s' in directive/ast.set - field", f.Kind())
		}
		return nil
	}

	return fmt.Errorf("%T did not have method or field %s", x, field)
}
