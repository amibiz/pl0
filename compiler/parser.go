package compiler

import (
	"io"

	"pl0/compiler/token"
)

var level int // Lexical level

// match checks for a specific token.
func match(want token.Token) {
	if tok == want {
		next()
	} else {
		expected(want.String(), text)
	}
}

// checkIdent checks for an identifier literal.
func checkIdent() {
	if tok != token.IDENT {
		expected("identifier", text)
	}
}

// ParseAndTranslate parses and translates a program.
func ParseAndTranslate(in io.Reader, out io.Writer, name string) {
	initScanner(in)
	initCode(out)

	header(name)
	prolog()
	next()
	mainBlock()
	match(token.PERIOD)
	epilog()
}

// mainBlock parses and translates the top level main procedure.
func mainBlock() {
	level = 0
	block("MAIN")
	allocStatic(universe)
}

// block parses and translates a block.
func block(name string) {
	nvar := 0
	if tok == token.CONST {
		match(token.CONST)
		constDecl()
		for tok == token.COMMA {
			match(token.COMMA)
			constDecl()
		}
		match(token.SEMICOLON)
	}
	if tok == token.VAR {
		match(token.VAR)
		nvar++
		varDecl(nvar)
		for tok == token.COMMA {
			match(token.COMMA)
			nvar++
			varDecl(nvar)
		}
		match(token.SEMICOLON)
	}
	for tok == token.PROCEDURE {
		match(token.PROCEDURE)
		procDecl()
		match(token.SEMICOLON)
	}
	if level == 0 {
		nvar = 0
	}
	procProlog(name, nvar)
	statement()
	procEpilog()
}

// constDecl parses a constant declaration.
func constDecl() {
	checkIdent()
	name := text
	next()
	match(token.EQL)
	val := number()
	obj := newObj(name, constCls)
	obj.lev = level
	obj.val = val
}

// varDecl parses a variable declaration.
func varDecl(pos int) {
	checkIdent()
	obj := newObj(text, varCls)
	next()
	obj.lev = level
	obj.pos = pos
}

// procDecl parses a procedure declaration.
func procDecl() {
	level++
	checkIdent()
	obj := newObj(text, procCls)
	next()
	obj.lev = level
	match(token.SEMICOLON)
	openScope()
	block(obj.name)
	obj.dsc = topScope.next
	closeScope()
	level--
}

// statement parses and translates a statement.
func statement() {
	switch tok {
	case token.IDENT:
		assignment()
	case token.CALL:
		callProc()
	case token.BEGIN:
		begin()
	case token.IF:
		doIf()
	case token.WHILE:
		while()
	case token.SEND:
		send()
	case token.RECV:
		receive()
	}
}

// send parses and translates a "!".
func send() {
	match(token.SEND)
	expression()
	printNumber()
}

// receive parses and translates a "?".
func receive() {
	match(token.RECV)
	inputNumber()
	checkIdent()
	obj := find(text)
	next()
	if obj.kind == varCls {
		storeVariable(obj, level)
	} else {
		report("cannot receive into " + obj.name + " (kind " + obj.kind.String() + ")")
	}
}

// callProc parses and translates a call statement.
func callProc() {
	match(token.CALL)
	checkIdent()
	obj := find(text)
	next()
	if obj.kind != procCls {
		report("cannot call non-procedure " + obj.name + " (kind " + obj.kind.String() + ")")
	}
	call(obj, level)
}

// begin parses and translates a begin statement.
func begin() {
	match(token.BEGIN)
	statement()
	for tok == token.SEMICOLON {
		match(token.SEMICOLON)
		statement()
	}
	match(token.END)
}

// assignment parses and translates an assignment statement.
func assignment() {
	checkIdent()
	obj := find(text)
	next()
	match(token.BECOMES)
	expression()
	if obj.kind == varCls {
		storeVariable(obj, level)
	} else {
		report("cannot assign to " + obj.name + " (kind " + obj.kind.String() + ")")
	}
}

// doIf parses and translates an if statement.
func doIf() {
	match(token.IF)
	l1 := newLabel()
	condition()
	branchFalse(l1)
	match(token.THEN)
	statement()
	postLabel(l1)
}

// while parses and translates a while statement.
func while() {
	match(token.WHILE)
	l1 := newLabel()
	l2 := newLabel()
	postLabel(l1)
	condition()
	branchFalse(l2)
	match(token.DO)
	statement()
	branch(l1)
	postLabel(l2)
}

// condition parses and translates a condition.
func condition() {
	if tok == token.ODD {
		odd()
	} else {
		relation()
	}
}

// relation parses and translates a relation.
func relation() {
	expression()
	if tok.IsRelop() {
		push()
		switch tok {
		case token.EQL:
			equals()
		case token.NEQ:
			notEquals()
		case token.LSS:
			less()
		case token.LEQ:
			lessOrEqual()
		case token.GRT:
			greater()
		case token.GEQ:
			greaterOrEqual()
		}
	} else {
		expected("relation", text)
	}
}

// odd recognizes and translates a relational "odd parity".
func odd() {
	match(token.ODD)
	expression()
	testParity()
	setOdd()
}

// greaterOrEqual recognizes and translates a relational "greater than".
func greaterOrEqual() {
	match(token.GEQ)
	expression()
	popCompare()
	setGreaterOrEqual()
}

// greater recognizes and translates a relational "greater than".
func greater() {
	match(token.GRT)
	expression()
	popCompare()
	setGreater()
}

// lessOrEqual recognizes and translates a relational "less or equal".
func lessOrEqual() {
	match(token.LEQ)
	expression()
	popCompare()
	setLessOrEqual()
}

// less recognizes and translates a relational "less than".
func less() {
	match(token.LSS)
	expression()
	popCompare()
	setLess()
}

// notEquals recognizes and translates a relational "Not equals".
func notEquals() {
	match(token.NEQ)
	expression()
	popCompare()
	setNotEqual()
}

// equals recognizes and translates a relational "equals".
func equals() {
	match(token.EQL)
	expression()
	popCompare()
	setEqual()
}

// expression parses and translats a maths expression.
func expression() {
	signedTerm()
	for tok.IsAddop() {
		switch tok {
		case token.PLUS:
			add()
		case token.MINUS:
			subtract()
		}
	}
}

// signedTerm parses and translates a maths term with optional leading sign.
func signedTerm() {
	sign := tok
	if tok.IsAddop() {
		next()
	}
	term()
	if sign == token.MINUS {
		negate()
	}
}

// add parses and translates an addition operation.
func add() {
	match(token.PLUS)
	push()
	term()
	popAdd()
}

// subtract parses and translates a subtraction operation.
func subtract() {
	match(token.MINUS)
	push()
	term()
	popSub()
	negate()
}

// term parses and translates  a maths term.
func term() {
	factor()
	for tok.IsMulop() {
		switch tok {
		case token.TIMES:
			multiply()
		case token.DIV:
			divide()
		}
	}
}

// multiply parses and translates a multiply.
func multiply() {
	match(token.TIMES)
	push()
	factor()
	popMul()
}

// divide parses and translates a divide.
func divide() {
	match(token.DIV)
	push()
	factor()
	popDiv()
}

// factor parses and translates a maths factor.
func factor() {
	if tok == token.LPAREN {
		match(token.LPAREN)
		expression()
		match(token.RPARAN)
	} else if tok == token.NUMBER {
		loadConstant(number())
	} else if tok == token.IDENT {
		checkIdent()
		obj := find(text)
		next()
		if obj.kind == varCls {
			loadVariable(obj, level)
		} else if obj.kind == constCls {
			loadConstant(obj.val)
		} else {
			report("cannot use " + obj.name + " (kind " + obj.kind.String() + ") in expression")
		}
	} else {
		expected("expression", text)
	}
}

// number parses and returns a number literal.
func number() string {
	if tok != token.NUMBER {
		expected("number", text)
	}
	value := text
	next()
	return value
}
