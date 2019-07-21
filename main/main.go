package main

import (
	"log"

	"github.com/dianelooney/directive/format"

	. "github.com/dianelooney/directive/ast"
)

func main() {
	const doc = `
	TimeTop 6
	TimeBot 8
	Tempo 150
	
	Kit {
		Name "x"
		Volume 0.1
		[Sample "808s_2" "hihats_1"]
		Loop {
			Measure {
				[Pulse 	1	2	3			]
			}
		}
	}
	
	Wave {
		Name "carrot"
		Volume 0.01
		BaseFreq 440
		Chord "major"
		Pattern "sin"
		Vibrato 1.05

		A { b "c" };
	
		Loop {
			Measure {
				[Note 	1]
				[Len  	6]
				[Pulse	1]
			}
		}
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

	format.PrettyPrint(d)
}
