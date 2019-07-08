package ast_test

import "testing"

import . "github.com/dianelooney/directive/ast"

type Doc struct {
	Time  string
	Tempo float64
	Kits  []*Kit
}

func (d *Doc) Kit() *Kit {
	i := &Kit{}
	d.Kits = append(d.Kits, i)
	return i
}

type Kit struct {
	Sample string
	Loops  []*Loop
}

func (k *Kit) Loop() *Loop {
	l := &Loop{}
	k.Loops = append(k.Loops, l)
	return l
}

type Loop struct {
	Measures []*Measure
}

func (l *Loop) Measure() *Measure {
	m := &Measure{}
	l.Measures = append(l.Measures, m)
	return m
}

type Measure struct {
	Pulses []float64
}

func (m *Measure) Pulse(t float64) {
	m.Pulses = append(m.Pulses, t)
}
func TestParseDocument(t *testing.T) {
	const doc = `
	Time "4/4"
	Tempo 120
	
	Kit {
		Sample "bass_1"
		Loop {
			Measure {
				[Pulse 1 2 3 4]
			}
		}
	}
	
	Kit {
		Sample "snare_1"
		Loop {
			Measure {
				[Pulse 2.33 2.66 4.33 4.66]
			}
		}
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

	document := Doc{}
	d.Execute(&document)
	if document.Time != "4/4" {
		t.Errorf("Parse returned an incorrect string argument. Expected '4/4' but got '%v'", document.Time)
	}
}
