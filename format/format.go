package format

import (
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dianelooney/directive/ast"
)

func Prettify(n ast.Node, wr io.Writer) {
	w := new(tabwriter.Writer)
	w.Init(wr, 6, 0, 1, ' ', 0)
	print(w, n, 0)
	w.Flush()
}

func PrettyPrint(n ast.Node) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 6, 0, 1, ' ', 0)
	print(w, n, 0)
	w.Flush()
}

func print(w *tabwriter.Writer, n ast.Node, i int) {
	indent := strings.Repeat("\t", i)

	switch v := n.(type) {
	case ast.Whitespace:
		w.Write([]byte{'\n'})
	case *ast.Directive:
		if _, ok := v.Value.(*ast.Object); ok {
			if v.HasSemi {
				w.Write([]byte(indent + v.Identifier + "\t{"))
				printSingle(w, v.Value, i)
				w.Write([]byte("};\n"))
			} else {
				w.Write([]byte(indent + v.Identifier + "\t{\n"))
				print(w, v.Value, i)
				w.Write([]byte(indent + "}\n"))
			}
		} else {
			w.Write([]byte(indent + v.Identifier + "\t" + v.Value.Text() + "\n"))
		}
	case *ast.RepeatedDirective:
		s := indent + "[" + v.Identifier
		for _, x := range v.Values {
			s += "\t" + x.Text()
		}
		s += "]\n"
		w.Write([]byte(s))
	case *ast.Document:
		for _, d := range v.Directives {
			print(w, d, i)
		}
	case *ast.Object:
		for _, d := range v.Directives {
			print(w, d, i+1)
		}
	}
}
func printSingle(w *tabwriter.Writer, n ast.Node, i int) {
	indent := strings.Repeat("\t", i)

	switch v := n.(type) {
	case ast.Whitespace:
	case *ast.Directive:
		if _, ok := v.Value.(*ast.Object); ok {
			if v.HasSemi {
				w.Write([]byte(v.Identifier + " {"))
				printSingle(w, v.Value, i)
				w.Write([]byte("}"))
			}
		} else {
			w.Write([]byte(" " + v.Identifier + "\t" + v.Value.Text() + " "))
		}
	case *ast.RepeatedDirective:
		s := indent + "[" + v.Identifier
		for _, x := range v.Values {
			s += "\t" + x.Text()
		}
		s += "]\t"
		w.Write([]byte(s))
	case *ast.Document:
		for _, d := range v.Directives {
			printSingle(w, d, i)
		}
	case *ast.Object:
		for _, d := range v.Directives {
			printSingle(w, d, i+1)
		}
	}
}
