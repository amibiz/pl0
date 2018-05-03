package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pl0/compiler"
)

var s = flag.Bool("S", false, "only output assembly")
var o = flag.String("o", "", "resulting executable name")

var pl0root = "/usr/local/pl0"

func init() {
	if custom := os.Getenv("PL0ROOT"); custom != "" {
		pl0root = custom
	}
	if fi, err := os.Stat(pl0root); err != nil || !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "pl0: cannot find PL0ROOT directory: %v\n", pl0root)
		os.Exit(2)
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "pl0: multiple pl0 files given\n")
		os.Exit(2)
	}
	if len(args) == 0 || !strings.HasSuffix(args[0], ".pl0") {
		fmt.Fprintf(os.Stderr, "pl0: no pl0 file given\n")
		os.Exit(2)
	}

	compile(args[0])
}

func compile(pl0file string) {
	// Create the intermediate assembly output file
	asmfile, err := ioutil.TempFile("", "pl0__")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
	defer os.Remove(asmfile.Name())

	srcfile, err := os.Open(pl0file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	progname := strings.TrimSuffix(filepath.Base(srcfile.Name()), ".pl0")
	if *o != "" {
		progname = *o
	}

	// Compile
	compiler.ParseAndTranslate(srcfile, asmfile, progname)

	if *s {
		asmfile.Seek(0, 0) // Rewind to the beginning
		if _, err := io.Copy(os.Stdout, asmfile); err != nil {
			log.Fatal(err)
		}
		return
	}

	runtime := filepath.Join(pl0root, "include", "runtime.asm")
	nasmpath := filepath.Join(pl0root, "bin", "asm")

	// Create object file
	objpath := asmfile.Name() + ".o"
	assembler := exec.Command(nasmpath, "-p", runtime, "-f", "macho32", "-o", objpath, asmfile.Name())
	assembler.Stderr = os.Stderr
	assembler.Stdout = os.Stdout
	if err := assembler.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer os.Remove(objpath)

	// Create binary executable
	linker := exec.Command("ld", "-e", "start", "-o", progname, objpath)
	linker.Stderr = os.Stderr
	linker.Stdout = os.Stdout
	if err := linker.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
