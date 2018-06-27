package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"pl0/compiler/ast"
	"pl0/compiler/token"
)

func TestParseExpr(t *testing.T) {
	tests := []struct {
		in   string
		env  Env
		want int
	}{
		{"", nil, 0},
		{"10", nil, 10},
		{"x", Env{"x": 5}, 5},
		{"(y)", Env{"y": 1}, 1},
		{"(((z)))", Env{"z": 1}, 1},
		{"-7", nil, -7},
		{"+3 * (-7)", nil, -21},
		{"3 * 5", nil, 15},
		{"(10 / x) * 4", Env{"x": 5}, 8},
		{"1 + 2", nil, 3},
		{"9 - (5 + 3)", nil, 1},
		{"z * (x / 2) - (y + 3)", Env{"z": 9, "x": 6, "y": 4}, 20},
	}
	for _, tt := range tests {
		x := ParseExpr(strings.NewReader(tt.in))
		got := Eval(x, tt.env)
		if got != tt.want {
			t.Errorf("ParseExpr(%q): got %d, want %d", tt.in, got, tt.want)
		}
	}
}

// Env maps identifiers to number values.
type Env map[string]int

// Eval is a helper for testing expressions. Given an expression in abstract
// form, it evaluates the result. Identifiers are supplied in the Env map.
func Eval(x ast.Expr, e Env) int {
	switch n := x.(type) {
	case *ast.Ident:
		return e[n.Name]

	case *ast.Number:
		val, _ := strconv.Atoi(n.Value)
		return val

	case *ast.UnaryExpr:
		e1 := Eval(n.X, e)
		switch n.Op {
		case token.PLUS:
			return +e1
		case token.MINUS:
			return -e1
		}
		panic(fmt.Sprintf("unsupported unary operator: %q", n.Op))

	case *ast.BinaryExpr:
		e1 := Eval(n.X, e)
		e2 := Eval(n.Y, e)
		switch n.Op {
		case token.PLUS:
			return e1 + e2
		case token.MINUS:
			return e1 - e2
		case token.TIMES:
			return e1 * e2
		case token.DIV:
			return e1 / e2
		}
		panic(fmt.Sprintf("unsupported binary operator: %q", n.Op))
	}
	return 0
}
