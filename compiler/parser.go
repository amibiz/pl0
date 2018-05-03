package compiler

import "io"

var level int // Lexical level

// match checks for a specific token.
func match(want Token) {
	if token == want {
		next()
	} else {
		expected(want.String(), text)
	}
}

// checkIdent checks for an identifier literal.
func checkIdent() {
	if token != IDENT {
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
	match(PERIOD)
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
	if token == CONST {
		match(CONST)
		constDecl()
		for token == COMMA {
			match(COMMA)
			constDecl()
		}
		match(SEMICOLON)
	}
	if token == VAR {
		match(VAR)
		nvar++
		varDecl(nvar)
		for token == COMMA {
			match(COMMA)
			nvar++
			varDecl(nvar)
		}
		match(SEMICOLON)
	}
	for token == PROCEDURE {
		match(PROCEDURE)
		procDecl()
		match(SEMICOLON)
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
	match(EQL)
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
	match(SEMICOLON)
	openScope()
	block(obj.name)
	obj.dsc = topScope.next
	closeScope()
	level--
}

// statement parses and translates a statement.
func statement() {
	switch token {
	case IDENT:
		assignment()
	case CALL:
		callProc()
	case BEGIN:
		begin()
	case IF:
		doIf()
	case WHILE:
		while()
	case SEND:
		send()
	case RECV:
		receive()
	}
}

// send parses and translates a "!".
func send() {
	match(SEND)
	expression()
	printNumber()
}

// receive parses and translates a "?".
func receive() {
	match(RECV)
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
	match(CALL)
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
	match(BEGIN)
	statement()
	for token == SEMICOLON {
		match(SEMICOLON)
		statement()
	}
	match(END)
}

// assignment parses and translates an assignment statement.
func assignment() {
	checkIdent()
	obj := find(text)
	next()
	match(BECOMES)
	expression()
	if obj.kind == varCls {
		storeVariable(obj, level)
	} else {
		report("cannot assign to " + obj.name + " (kind " + obj.kind.String() + ")")
	}
}

// doIf parses and translates an if statement.
func doIf() {
	match(IF)
	l1 := newLabel()
	condition()
	branchFalse(l1)
	match(THEN)
	statement()
	postLabel(l1)
}

// while parses and translates a while statement.
func while() {
	match(WHILE)
	l1 := newLabel()
	l2 := newLabel()
	postLabel(l1)
	condition()
	branchFalse(l2)
	match(DO)
	statement()
	branch(l1)
	postLabel(l2)
}

// condition parses and translates a condition.
func condition() {
	if token == ODD {
		odd()
	} else {
		relation()
	}
}

// relation parses and translates a relation.
func relation() {
	expression()
	if token.IsRelop() {
		push()
		switch token {
		case EQL:
			equals()
		case NEQ:
			notEquals()
		case LSS:
			less()
		case LEQ:
			lessOrEqual()
		case GRT:
			greater()
		case GEQ:
			greaterOrEqual()
		}
	} else {
		expected("relation", text)
	}
}

// odd recognizes and translates a relational "odd parity".
func odd() {
	match(ODD)
	expression()
	testParity()
	setOdd()
}

// greaterOrEqual recognizes and translates a relational "greater than".
func greaterOrEqual() {
	match(GEQ)
	expression()
	popCompare()
	setGreaterOrEqual()
}

// greater recognizes and translates a relational "greater than".
func greater() {
	match(GRT)
	expression()
	popCompare()
	setGreater()
}

// lessOrEqual recognizes and translates a relational "less or equal".
func lessOrEqual() {
	match(LEQ)
	expression()
	popCompare()
	setLessOrEqual()
}

// less recognizes and translates a relational "less than".
func less() {
	match(LSS)
	expression()
	popCompare()
	setLess()
}

// notEquals recognizes and translates a relational "Not equals".
func notEquals() {
	match(NEQ)
	expression()
	popCompare()
	setNotEqual()
}

// equals recognizes and translates a relational "equals".
func equals() {
	match(EQL)
	expression()
	popCompare()
	setEqual()
}

// expression parses and translats a maths expression.
func expression() {
	signedTerm()
	for token.IsAddop() {
		switch token {
		case PLUS:
			add()
		case MINUS:
			subtract()
		}
	}
}

// signedTerm parses and translates a maths term with optional leading sign.
func signedTerm() {
	sign := token
	if token.IsAddop() {
		next()
	}
	term()
	if sign == MINUS {
		negate()
	}
}

// add parses and translates an addition operation.
func add() {
	match(PLUS)
	push()
	term()
	popAdd()
}

// subtract parses and translates a subtraction operation.
func subtract() {
	match(MINUS)
	push()
	term()
	popSub()
	negate()
}

// term parses and translates  a maths term.
func term() {
	factor()
	for token.IsMulop() {
		switch token {
		case TIMES:
			multiply()
		case DIV:
			divide()
		}
	}
}

// multiply parses and translates a multiply.
func multiply() {
	match(TIMES)
	push()
	factor()
	popMul()
}

// divide parses and translates a divide.
func divide() {
	match(DIV)
	push()
	factor()
	popDiv()
}

// factor parses and translates a maths factor.
func factor() {
	if token == LPAREN {
		match(LPAREN)
		expression()
		match(RPARAN)
	} else if token == NUMBER {
		loadConstant(number())
	} else if token == IDENT {
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
	if token != NUMBER {
		expected("number", text)
	}
	value := text
	next()
	return value
}
