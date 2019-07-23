package eval

import (
	"fmt"
	"strconv"
	"strings"
)

type Node interface {
	Evaluate() []float64
}

type Group struct {
	LParen Node
	Child  Node
	RParen Node
}

func (n *Group) Evaluate() (out []float64) {
	return n.Child.Evaluate()
}

func (n *Group) String() string {
	return fmt.Sprintf("(%s)", n.Child)
}

type Operator struct {
	Op  Token
	LHS Node
	RHS Node
}

func (n *Operator) Evaluate() (out []float64) {
	left := n.LHS.Evaluate()
	right := n.RHS.Evaluate()

	switch n.Op {
	case "+":
		i := 0
		out = make([]float64, len(left)*len(right))
		for _, x := range left {
			for _, y := range right {
				out[i] = x + y
				i++
			}
		}
		return
	case "-":
		i := 0
		out = make([]float64, len(left)*len(right))
		for _, x := range left {
			for _, y := range right {
				out[i] = x - y
				i++
			}
		}
		return
	case "%":
		mod := right[0]
		min := right[1]
		max := right[2]
		for _, v := range left {
			for v >= min {
				v -= mod
			}
			for v < min {
				v += mod
			}
			for ; v < max; v += mod {
				out = append(out, v)
			}
		}
		return out
	case "*":
		for _, count := range right {
			for _, x := range left {
				for i := 0; i < int(count); i++ {
					out = append(out, x)
				}
			}
		}
		return out
	}

	panic("operator " + string(n.Op) + " not supported in *Operator.Evaluate()")
}

func (n *Operator) String() string {
	return fmt.Sprintf("%s %s %s", n.LHS, n.Op, n.RHS)
}

type Value struct {
	Tok Token
}

func (n *Value) Evaluate() (out []float64) {
	if idx := strings.Index(string(n.Tok), "/"); idx >= 0 {
		lhs, _ := strconv.ParseFloat(string(n.Tok)[:idx], 64)
		rhs, _ := strconv.ParseFloat(string(n.Tok)[idx+1:], 64)
		return []float64{lhs / rhs}
	}
	f, _ := strconv.ParseFloat(string(n.Tok), 64)
	return []float64{f}
}

func (n *Value) String() string {
	return string(n.Tok)
}

type List struct {
	Nums []Node
}

func (n *List) Evaluate() (out []float64) {
	for _, num := range n.Nums {
		out = append(out, num.Evaluate()...)
	}
	return
}
