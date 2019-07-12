package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dianelooney/directive/ast"
	"github.com/dianelooney/directive/format"
)

func main() {
	files := os.Args[1:]
	for _, path := range files {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading '%s': %v\n", path, err)
			continue
		}
		p := ast.NewParser(data)
		d, err := p.Parse()
		if err != nil {
			fmt.Printf("Unable to parse '%s': %v", path, err)
			continue
		}

		var buf bytes.Buffer
		format.Prettify(d, &buf)
		err = ioutil.WriteFile(path, buf.Bytes(), 0666)
		if err != nil {
			fmt.Printf("Unable to write to file '%s': %v", path, err)
		}
	}
}
