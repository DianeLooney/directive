package main

import (
	"fmt"
	"log"

	. "github.com/dianelooney/directive/ast"
)

func main() {
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
		log.Fatalf("Parse returned an error: %v", err)
	}

	if d == nil {
		log.Fatalf("Parse returned a nil Document")
	}

	fmt.Printf("%s\n", d)
}
