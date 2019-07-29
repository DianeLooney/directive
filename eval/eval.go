package eval

import (
	"regexp"
	"strconv"
	"strings"
)

type Token string

func (t Token) Evaluate() (out []float64) {
	v, _ := strconv.ParseFloat(string(t), 64)
	return []float64{v}
}

func Eval(s string) (out []float64) {
	tkns := Tokenize(s)
	p := Parser{
		tkns: tkns,
	}
	return p.Parse().Evaluate()
}

var symbols = []string{
	`+`,
	`-`,
	`&`,
	`|`,
	`%`,
	//`<`,
	//`>`,
	`\`,
	//`…`,
	`(`,
	`)`,
	`*`,
}
var tkn = regexp.MustCompile(`(?:[0-9]+/[0-9]+)|(?:-?[0-9]+(?:\.[0-9]*)?)|x|\` + strings.Join(symbols, `|\`))

func Tokenize(s string) []Token {
	tkns := tkn.FindAllString(s, -1)
	out := make([]Token, len(tkns))
	for i, t := range tkns {
		out[i] = Token(t)
	}
	return out
}

type Parser struct {
	tkns []Token
}

func (p *Parser) Parse() Node {
	var nodes []Node
	for {
		if len(p.tkns) == 0 {
			break
		}
		t := p.tkns[0]
		if t == "(" {
			c := p.parseGroup()
			nodes = append(nodes, c)
		} else {
			tok := p.parseAny()
			nodes = append(nodes, tok)
		}
	}

	return compact(nodes)
}

func compact(in []Node) (out Node) {
	for _, ops := range order {
		for i, n := range in {
			t, ok := n.(Token)
			if !ok {
				continue
			}
			if !ops[string(t)] {
				continue
			}

			lhs := in[:i]
			rhs := in[i+1:]
			oper := Operator{
				Op:  t,
				LHS: compact(lhs),
				RHS: compact(rhs),
			}
			return &oper
		}
	}
	list := make([]Node, len(in))
	for i, n := range in {
		if tok, ok := n.(Token); ok {
			list[i] = &Value{Tok: tok}
		}
		if grp, ok := n.(*Group); ok {
			list[i] = grp
		}
	}
	return &List{Nums: list}
}

func (p *Parser) mustParse(s string) Token {
	if p.tkns[0] != Token(s) {
		panic("expected a " + s)
	}

	return p.parseAny()
}
func (p *Parser) parseAny() Token {
	if len(p.tkns) == 0 {
		panic("out of tokens")
	}
	out := p.tkns[0]
	p.tkns = p.tkns[1:]

	return out
}
func (p *Parser) parseGroup() Node {
	n := Group{}
	n.LParen = p.mustParse("(")
	var children []Token
	for {
		if p.tkns[0] == ")" {
			n.RParen = p.mustParse(")")
			break
		}
		children = append(children, p.parseAny())
	}
	p2 := Parser{}
	p2.tkns = children
	n.Child = p2.Parse()
	return &n
}
