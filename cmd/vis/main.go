package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"pl0/compiler"
	"pl0/compiler/ast"
	"pl0/compiler/token"
)

var version string

func usage() {
	fmt.Fprintf(os.Stderr, `usage: %s pl0file

Visualize the program comprising the named PL/0 source file.
A PL/0 source file is defined to be a file ending in a literal ".pl0" suffix.

version: %s

`,
		filepath.Base(os.Args[0]), version)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("vis: ")

	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "vis: multiple pl0 files given\n")
		os.Exit(2)
	}
	if len(args) == 0 || !strings.HasSuffix(args[0], ".pl0") {
		fmt.Fprintf(os.Stderr, "vis: no pl0 file given\n")
		os.Exit(2)
	}

	p, err := compiler.Parse(args[0], nil)
	if err != nil {
		log.Fatal(err)
	}
	dot(p)
}

func dot(n ast.Node) {
	switch n := n.(type) {
	case *ast.Program:
		fmt.Printf("digraph %q {\n", n.Name)
		printNode(n, "Program")
		printEdge(n, n.Main)
		dot(n.Main)
		fmt.Println("}")

	case *ast.Block:
		printNode(n, "Block")
		for _, c := range n.Consts {
			printEdge(n, c)
			dot(c)
		}
		for _, v := range n.Vars {
			printEdge(n, v)
			dot(v)
		}
		for _, p := range n.Procs {
			printEdge(n, p)
			dot(p)
		}
		printEdge(n, n.Body)
		dot(n.Body)

	case *ast.ConstDecl:
		printNode(n, "ConstDecl")
		printEdge(n, n.Name)
		printEdge(n, n.Value)
		dot(n.Name)
		dot(n.Value)

	case *ast.ProcDecl:
		printNode(n, "ProcDecl")
		printEdge(n, n.Name)
		printEdge(n, n.Block)
		dot(n.Name)
		dot(n.Block)

	case *ast.AssignStmt:
		printNode(n, ":=")
		printEdge(n, n.Lhs)
		printEdge(n, n.Rhs)
		dot(n.Lhs)
		dot(n.Rhs)

	case *ast.CallStmt:
		printNode(n, "CALL")
		printEdge(n, n.Proc)
		dot(n.Proc)

	case *ast.SendStmt:
		printNode(n, "!")
		printEdge(n, n.X)
		dot(n.X)

	case *ast.ReceiveStmt:
		printNode(n, "?")
		printEdge(n, n.Name)
		dot(n.Name)

	case *ast.BeginStmt:
		printNode(n, "Stmts")
		for _, s := range n.List {
			printEdge(n, s)
			dot(s)
		}

	case *ast.IfStmt:
		printNode(n, "IF")
		printEdge(n, n.Cond)
		printEdge(n, n.Body)
		dot(n.Cond)
		dot(n.Body)

	case *ast.WhileStmt:
		printNode(n, "WHILE")
		printEdge(n, n.Cond)
		printEdge(n, n.Body)
		dot(n.Cond)
		dot(n.Body)

	case *ast.OddCond:
		printNode(n, "ODD")
		printEdge(n, n.X)
		dot(n.X)

	case *ast.RelCond:
		printNode(n, n.Op.String())
		printEdge(n, n.X)
		printEdge(n, n.Y)
		dot(n.X)
		dot(n.Y)

	case *ast.Ident:
		printNode(n, n.Name)

	case *ast.Number:
		printNode(n, n.Value)

	case *ast.UnaryExpr:
		switch n.Op {
		case token.PLUS, token.MINUS:
			printNode(n, n.Op.String())
			printEdge(n, n.X)
			dot(n.X)
		default:
			panic(fmt.Sprintf("unsupported unary operator: %q", n.Op))
		}

	case *ast.BinaryExpr:
		switch n.Op {
		case token.PLUS, token.MINUS, token.TIMES, token.DIV:
			printNode(n, n.Op.String())
			printEdge(n, n.X)
			printEdge(n, n.Y)
			dot(n.X)
			dot(n.Y)
		default:
			panic(fmt.Sprintf("unsupported binary operator: %q", n.Op))
		}

	default:
		panic(fmt.Sprintf("unsupported node: %T", n))
	}
}

func printNode(n ast.Node, display string) {
	fmt.Printf("\t\"%p\" [label=%q];\n", n, display)
}

func printEdge(x, y ast.Node) {
	fmt.Printf("\t\"%p\" -> \"%p\";\n", x, y)
}
