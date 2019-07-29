package eval_test

import (
	"testing"

	"github.com/dianelooney/directive/eval"
)

type testCase struct {
	str      string
	expected []float64
}

func (c testCase) Test(t *testing.T) {
	actual := eval.Eval(c.str)
	expected := c.expected

	if len(actual) != len(expected) {
		t.Errorf("Expected %v to match %v", actual, c.expected)
	}
	for i, v := range actual {
		if v-expected[i] > 0.000001 {
			t.Errorf("Expected %v to match %v", actual, c.expected)
			return
		}
	}
}

func TestEval_1(t *testing.T) {
	testCase{
		str:      `0 1 2`,
		expected: []float64{0, 1, 2},
	}.Test(t)
}

func TestEval_Add(t *testing.T) {
	testCase{
		str:      `1 + 2`,
		expected: []float64{3},
	}.Test(t)
	testCase{
		str:      `1 2 + 3 4`,
		expected: []float64{4, 5, 5, 6},
	}.Test(t)
}
func TestEval_Sub(t *testing.T) {
	testCase{
		str:      `1 - 2`,
		expected: []float64{-1},
	}.Test(t)
	testCase{
		str:      `1 2 - 3 4`,
		expected: []float64{-2, -3, -1, -2},
	}.Test(t)
}
func TestEval_Mod(t *testing.T) {
	testCase{
		str:      `0 % 1 0 8`,
		expected: []float64{0, 1, 2, 3, 4, 5, 6, 7},
	}.Test(t)
	testCase{
		str:      `0 % 0.33 0 2`,
		expected: []float64{0, 0.33, 0.66, 0.99, 1.32, 1.65, 1.98},
	}.Test(t)
	testCase{
		str:      `1/3 2/3 % 1 0 2`,
		expected: []float64{1.0 / 3, 4.0 / 3, 2.0 / 3, 5.0 / 3},
	}.Test(t)
}
func TestEval_Rep(t *testing.T) {
	testCase{
		str:      `0 * 8`,
		expected: []float64{0, 0, 0, 0, 0, 0, 0, 0},
	}.Test(t)
	testCase{
		str:      `4 5 6 * 1 2 3`,
		expected: []float64{4, 5, 6, 4, 4, 5, 5, 6, 6, 4, 4, 4, 5, 5, 5, 6, 6, 6},
	}.Test(t)
}

func TestEval_Thing(t *testing.T) {
	testCase{
		str:      `4 3 4 2 + 0 -1`,
		expected: []float64{4, 3, 4, 2, 3, 2, 3, 1},
	}.Test(t)
}

func TestTokenize(t *testing.T) {
	eval.Tokenize(`1(2 3 %4)`)
}
