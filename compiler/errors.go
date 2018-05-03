package compiler

import (
	"fmt"
	"os"
)

// report writes error message and halt.
func report(s string) {
	fmt.Fprintf(os.Stderr, "error:%d:%s\n", lineno, s)
	os.Exit(1)
}

// undefined reports an undefined identifier.
func undefined(ident string) {
	report("undefined identifier " + ident)
}

// duplicate reports a duplicate identifier.
func duplicate(ident string) {
	report("duplicate identifier " + ident)
}

// expected reports what was expected.
func expected(want, got string) {
	report("unexpected " + got + ", expecting " + want)
}
