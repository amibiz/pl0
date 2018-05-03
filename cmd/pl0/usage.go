package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var version string

func usage() {
	fmt.Fprintf(os.Stderr, `usage: %s [-o output] [flags] pl0file

Compile the program comprising the named PL/0 source file.
A PL/0 source file is defined to be a file ending in a literal ".pl0" suffix.

The resulting executable is written to an output file named after the source
file (e.g., 'pl0 primes.pl0' writes 'primes').

The -o flag forces the compiler to write the resulting executable
to the named output file, instead of the default behavior described
in the last paragraph.

The -S flag instructs the compiler to only output the assembly used
to create the resulting executable file (not including any internal
runtime code). In this case, the -o flag is ignored if provided.

version: %s

`,
		filepath.Base(os.Args[0]), version)
}
