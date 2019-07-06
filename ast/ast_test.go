package ast_test

import "testing"

import . "github.com/dianelooney/directive/ast"

func TestParseDocument(t *testing.T) {
	const doc = `
		name "something"
		other_name "something else"
		version "30"
		[author "diane" "john" "anonymous"]

		@note { freq "440"; duration "1.beat" }

		measure {
			[note {} {} {} {}]
		}
	`

	p := NewParser([]byte(doc))
	d, err := p.Parse()
	if err != nil {
		t.Errorf("Parse returned an error: %v", err)
	}

	if d == nil {
		t.Errorf("Parse returned a nil Document")
	}
}
