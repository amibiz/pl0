package compiler

import (
	"io"
	"os"

	"pl0/compiler/ast"
	"pl0/compiler/token"
)

// match checks for a specific token.
func match(want token.Token) {
	if tok == want {
		next()
	} else {
		expected(want.String(), text)
	}
}

// ParseAndTranslate parses and translates a program.
func ParseAndTranslate(in io.Reader, out io.Writer, name string) {
	prog, err := Parse(name, in)
	if err != nil {
		report("failed to parse program: " + err.Error())
	}
	gen(prog, out)
}

func ParseExpr(src io.Reader) ast.Expr {
	initScanner(src)
	next()
	if tok == token.EOF {
		return nil
	}
	return parseExpr()
}

func Parse(filename string, src io.Reader) (*ast.Program, error) {
	if src == nil {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		src = f
	}
	initScanner(src)
	next()
	b := parseBlock()
	match(token.PERIOD)
	return &ast.Program{Name: filename, Main: b}, nil
}

func parseBlock() *ast.Block {
	b := new(ast.Block)

	if tok == token.CONST {
		match(token.CONST)
		c := make([]*ast.ConstDecl, 1)
		c[0] = parseConstDecl()
		for tok == token.COMMA {
			match(token.COMMA)
			c = append(c, parseConstDecl())
		}
		match(token.SEMICOLON)
		b.Consts = c
	}
	if tok == token.VAR {
		match(token.VAR)
		v := make([]*ast.Ident, 1)
		v[0] = parseIdent()
		for tok == token.COMMA {
			match(token.COMMA)
			v = append(v, parseIdent())
		}
		match(token.SEMICOLON)
		b.Vars = v
	}
	var p []*ast.ProcDecl
	for tok == token.PROCEDURE {
		p = append(p, parseProc())
	}
	b.Procs = p

	b.Body = parseStmt()
	return b
}

func parseConstDecl() *ast.ConstDecl {
	name := parseIdent()
	match(token.EQL)
	return &ast.ConstDecl{Name: name, Value: parseNumber()}
}

func parseProc() *ast.ProcDecl {
	match(token.PROCEDURE)
	name := parseIdent()
	match(token.SEMICOLON)
	block := parseBlock()
	match(token.SEMICOLON)
	return &ast.ProcDecl{Name: name, Block: block}
}

func parseStmt() ast.Stmt {
	switch tok {
	case token.IDENT:
		return parseAssign()
	case token.CALL:
		return parseCall()
	case token.SEND:
		return parseSend()
	case token.RECV:
		return parseReceive()
	case token.BEGIN:
		return parseBegin()
	case token.IF:
		return parseIf()
	case token.WHILE:
		return parseWhile()
	}
	return nil
}

func parseAssign() *ast.AssignStmt {
	i := parseIdent()
	match(token.BECOMES)
	x := parseExpr()
	return &ast.AssignStmt{Lhs: i, Rhs: x}
}

func parseCall() *ast.CallStmt {
	match(token.CALL)
	return &ast.CallStmt{Proc: parseIdent()}
}

func parseSend() *ast.SendStmt {
	match(token.SEND)
	return &ast.SendStmt{X: parseExpr()}
}

func parseReceive() *ast.ReceiveStmt {
	match(token.RECV)
	return &ast.ReceiveStmt{Name: parseIdent()}
}

func parseBegin() *ast.BeginStmt {
	match(token.BEGIN)
	s := make([]ast.Stmt, 1)
	s[0] = parseStmt()
	for tok == token.SEMICOLON {
		match(token.SEMICOLON)
		s = append(s, parseStmt())
	}
	match(token.END)
	return &ast.BeginStmt{List: s}
}

func parseIf() *ast.IfStmt {
	match(token.IF)
	c := parseCond()
	match(token.THEN)
	s := parseStmt()
	return &ast.IfStmt{Cond: c, Body: s}
}

func parseWhile() *ast.WhileStmt {
	match(token.WHILE)
	c := parseCond()
	match(token.DO)
	s := parseStmt()
	return &ast.WhileStmt{Cond: c, Body: s}
}

func parseCond() ast.Cond {
	if tok == token.ODD {
		match(token.ODD)
		return &ast.OddCond{X: parseExpr()}
	}
	return parseRel()
}

func parseRel() *ast.RelCond {
	x := parseExpr()
	if !tok.IsRelop() {
		expected("relation", text)
		return nil
	}
	op := tok
	next()
	return &ast.RelCond{X: x, Op: op, Y: parseExpr()}
}

func parseExpr() ast.Expr {
	sign := tok
	if sign.IsAddop() {
		next()
	}
	x := parseTerm()
	if sign.IsAddop() {
		x = &ast.UnaryExpr{X: x, Op: sign}
	}
	if tok.IsAddop() {
		op := tok
		next()
		return &ast.BinaryExpr{X: x, Op: op, Y: parseExpr()}
	}
	return x
}

func parseTerm() ast.Expr {
	x := parseFact()
	if tok.IsMulop() {
		op := tok
		next()
		return &ast.BinaryExpr{X: x, Op: op, Y: parseTerm()}
	}
	return x
}

func parseFact() ast.Expr {
	switch tok {
	case token.IDENT:
		return parseIdent()
	case token.NUMBER:
		return parseNumber()
	case token.LPAREN:
		match(token.LPAREN)
		x := parseExpr()
		match(token.RPARAN)
		return x
	default:
		expected("expression", text)
	}
	return nil
}

func parseIdent() *ast.Ident {
	if tok != token.IDENT {
		expected("identifier", text)
	}
	i := &ast.Ident{Name: text}
	next()
	return i
}

func parseNumber() *ast.Number {
	if tok != token.NUMBER {
		expected("number", text)
	}
	n := &ast.Number{Value: text}
	next()
	return n
}
