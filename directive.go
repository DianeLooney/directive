package directive

import (
	"fmt"

	"github.com/dianelooney/directive/ast"
)

type Executer interface {
	Execute(target interface{}) (err error)
}

func Execute(data []byte, target interface{}) (err error) {
	e, err := Prepare(data)
	if err != nil {
		return err
	}
	return e.Execute(target)
}

func Prepare(data []byte) (e Executer, err error) {
	p := ast.NewParser(data)
	doc, err := p.Parse()
	if err != nil {
		return nil, err
	}

	d, ok := doc.(*ast.Document)
	if !ok {
		return nil, fmt.Errorf("directive internal error: Parse didn't return a document")
	}

	return exeggutor{d}, nil
}

type exeggutor struct {
	doc *ast.Document
}

func (e exeggutor) Execute(target interface{}) (err error) {
	return e.Execute(target)
}
