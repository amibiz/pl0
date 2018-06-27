package compiler

import (
	"fmt"
	"io"
	"pl0/compiler/ast"
	"pl0/compiler/token"
	"strconv"
)

var (
	out     io.Writer // Output stream
	labelno int       // Label Counter
	level   int       // Lexical level
)

// gen takes a program in abstract form and generates code suitable for use
// by an assembler.
func gen(prog *ast.Program, w io.Writer) {
	out = w

	header(prog.Name)
	prolog()
	genMain(prog.Main)
	epilog()
}

// genMain emits code for the main program node.
func genMain(b *ast.Block) {
	level = 0
	genBlock("MAIN", b)
	allocStatic(universe)
}

// genBlock emits code for a block node.
func genBlock(name string, b *ast.Block) {
	for _, c := range b.Consts {
		obj := newObj(c.Name.Name, constCls)
		obj.lev = level
		obj.val = c.Value.Value
	}
	for p, v := range b.Vars {
		obj := newObj(v.Name, varCls)
		obj.lev = level
		obj.pos = p + 1
	}
	for _, p := range b.Procs {
		level++
		obj := newObj(p.Name.Name, procCls)
		obj.lev = level
		openScope()
		genBlock(obj.name, p.Block)
		obj.dsc = topScope.next
		closeScope()
		level--
	}
	procProlog(name, len(b.Vars))
	genStmt(b.Body)
	procEpilog()
}

// genStmt emits code for the various statement nodes.
func genStmt(s ast.Stmt) {
	switch s := s.(type) {
	case *ast.AssignStmt:
		obj := find(s.Lhs.Name)
		genExpr(s.Rhs)
		if obj.kind == varCls {
			storeVariable(obj, level)
		} else {
			report("cannot assign to " + obj.name + " (kind " + obj.kind.String() + ")")
		}

	case *ast.CallStmt:
		obj := find(s.Proc.Name)
		if obj.kind != procCls {
			report("cannot call non-procedure " + obj.name + " (kind " + obj.kind.String() + ")")
		}
		call(obj, level)

	case *ast.BeginStmt:
		for _, stmt := range s.List {
			genStmt(stmt)
		}

	case *ast.IfStmt:
		l1 := newLabel()
		genCond(s.Cond)
		branchFalse(l1)
		genStmt(s.Body)
		postLabel(l1)

	case *ast.WhileStmt:
		l1 := newLabel()
		l2 := newLabel()
		postLabel(l1)
		genCond(s.Cond)
		branchFalse(l2)
		genStmt(s.Body)
		branch(l1)
		postLabel(l2)

	case *ast.SendStmt:
		genExpr(s.X)
		printNumber()

	case *ast.ReceiveStmt:
		inputNumber()
		obj := find(s.Name.Name)
		if obj.kind == varCls {
			storeVariable(obj, level)
		} else {
			report("cannot receive into " + obj.name + " (kind " + obj.kind.String() + ")")
		}
	}
}

// genCond emits code for the various conditions nodes.
func genCond(c ast.Cond) {
	switch c := c.(type) {
	case *ast.OddCond:
		genExpr(c.X)
		testParity()
		setOdd()

	case *ast.RelCond:
		genExpr(c.X)
		push()
		genExpr(c.Y)
		popCompare()
		switch c.Op {
		case token.EQL:
			setEqual()
		case token.NEQ:
			setNotEqual()
		case token.LSS:
			setLess()
		case token.LEQ:
			setLessOrEqual()
		case token.GRT:
			setGreater()
		case token.GEQ:
			setGreaterOrEqual()
		default:
			report(fmt.Sprintf("unsupported relation operator: %q", c.Op))
		}
	}
}

// genExpr emits code for the various expression nodes.
func genExpr(x ast.Expr) {
	switch x := x.(type) {
	case *ast.UnaryExpr:
		genExpr(x.X)
		switch x.Op {
		case token.PLUS: // Noop case
		case token.MINUS:
			negate()
		default:
			report(fmt.Sprintf("unsupported unary operator: %q", x.Op))
		}

	case *ast.BinaryExpr:
		genExpr(x.X)
		push()
		genExpr(x.Y)
		switch x.Op {
		case token.PLUS:
			popAdd()
		case token.MINUS:
			popSub()
			negate()
		case token.TIMES:
			popMul()
		case token.DIV:
			popDiv()
		default:
			report(fmt.Sprintf("unsupported binary operator: %q", x.Op))
		}

	case *ast.Number:
		loadConstant(x.Value)

	case *ast.Ident:
		obj := find(x.Name)
		if obj.kind == varCls {
			loadVariable(obj, level)
		} else if obj.kind == constCls {
			loadConstant(obj.val)
		} else {
			report("cannot use " + obj.name + " (kind " + obj.kind.String() + ") in expression")
		}
	}
}

// write writes to the output stream.
func write(a ...interface{}) {
	fmt.Fprint(out, a...)
}

// writeln writes to the output stream, followed by a newline.
func writeln(a ...interface{}) {
	write(a...)
	fmt.Fprintln(out)
}

// emit emits an instruction.
func emit(s string) {
	write("\t", s)
}

// emit emits an instruction, followed by a newline.
func emitln(s string) {
	emit(s)
	writeln()
}

// newLabel generates a unique label.
func newLabel() string {
	defer func() { labelno++ }()
	return fmt.Sprintf("L%d", labelno)
}

// postLabel posts a label.
func postLabel(L string) {
	write(L + ":")
	writeln()
}

// header writes the program header info.
func header(name string) {
	writeln("; program: \"" + name + "\"")
	writeln(`;
; asm:   nasm
; os:    darwin
; arch:  386
;

global  start           ; must be declared for linker (ld)
`)
}

// prolog writes the program prolog.
func prolog() {
	writeln(`
section .text
start:                  ; tell linker entry point

	; call main program
	CALL MAIN

	; exit to operating system
	CALL EXIT

; compiled code starts here
;
`)
}

// epilog writes the program epilog.
func epilog() {
	writeln(`
; compiled code ends here
;`)
	writeln()
}

// call prepares and calls a procedure.
func call(proc *object, level int) {
	switch proc.lev {

	// Child
	case level + 1:
		emitln("PUSH EBP")

	// Peer
	case level:
		emitln("PUSH dword [EBP + 8]")

	// Ancestor
	default:
		walk(level - proc.lev)
		emitln("PUSH dword [EBX + 8]")
	}

	emitln("CALL " + proc.name)
	emitln("ADD ESP, 4") // Cleanup stack after return from procedure call
}

// doReturn from procedure call.
func doReturn() {
	emitln("RET")
	writeln("")
}

// write the prolog for a procedure.
func procProlog(name string, nvar int) {
	postLabel(name)
	emitln("PUSH EBP")
	emitln("MOV EBP, ESP")
	emitln("SUB ESP, " + strconv.Itoa(4*nvar))
	writeln("")
}

// write the epilog for a procedure.
func procEpilog() {
	writeln("")
	emitln("MOV ESP, EBP")
	emitln("POP EBP")
	doReturn()
}

// static prepends a variable with the "_" literal.
func static(name string) string {
	return "_" + name
}

// allocStatic allocates storage for a static variable.
func allocStatic(scope *object) {
	writeln()
	writeln()
	writeln(`section .data`)
	for x := scope.next; x != nil; x = x.next {
		if x.kind == varCls {
			writeln(static(x.name) + ": dd 0")
		}
	}
}

// loadConstant loads the primary register with a constant.
func loadConstant(number string) {
	emitln("MOV EAX, " + number)
}

// storeVariable stores the primary register with static, local or non-local variable.
func storeVariable(variable *object, level int) {
	offset := -4 * variable.pos

	switch variable.lev {

	// static
	case 0:
		emitln("MOV [" + static(variable.name) + "], EAX")

	// local
	case level:
		emitln("MOV [EBP + " + strconv.Itoa(offset) + "], EAX")

	// non-local
	default:
		walk(level - variable.lev)
		emitln("MOV [EBX + " + strconv.Itoa(offset) + "], EAX")
	}
}

// loadVariable loads the primary register with a static, local or non-local variable.
func loadVariable(variable *object, level int) {
	offset := -4 * variable.pos

	switch variable.lev {

	// static
	case 0:
		emitln("MOV EAX, [" + static(variable.name) + "]")

	// local
	case level:
		emitln("MOV EAX, [EBP + " + strconv.Itoa(offset) + "]")

	// non-local
	default:
		walk(level - variable.lev)
		emitln("MOV EAX, [EBX + " + strconv.Itoa(offset) + "]")
	}
}

// walk follows the static link chain n levels.
func walk(n int) {
	emitln("MOV EBX, [EBP + 8]")
	for i := 1; i < n; i++ {
		emitln("MOV EBX, [EBX + 8]")
	}
}

// push primary register to stack.
func push() {
	emitln("PUSH EAX")
}

// negate primary register.
func negate() {
	emitln("NEG EAX")
}

// popMul multiplies top-of-stack by primary register.
func popMul() {
	emitln("POP ECX")
	emitln("IMUL ECX")
}

// popDiv divides top of stack by primary register.
func popDiv() {
	emitln("MOV ECX, EAX")
	emitln("POP EAX")
	emitln("XOR EDX, EDX") // Clear EDX
	emitln("IDIV ECX")
}

// popAdd adds top-of-stack to primary register.
func popAdd() {
	emitln("POP EDX")
	emitln("ADD EAX, EDX")
}

// popSub subtracts top-of-stack from primary register.
func popSub() {
	emitln("POP EDX")
	emitln("SUB EAX, EDX")
}

// popCompare compares top of stack with primary register.
func popCompare() {
	emitln("POP EDX")
	emitln("CMP EDX, EAX")
}

// setEqual sets primary register if compare was "equal".
func setEqual() {
	emitln("CMOVE  EAX, [TRUE]")
	emitln("CMOVNE EAX, [FALSE]")
}

// setNotEqual sets primary register if compare was "not equal".
func setNotEqual() {
	emitln("CMOVNE EAX, [TRUE]")
	emitln("CMOVE  EAX, [FALSE]")
}

// setLess sets primary register if compare was "less than".
func setLess() {
	emitln("CMOVL  EAX, [TRUE]")
	emitln("CMOVGE EAX, [FALSE]")
}

// setGreater sets primary register if compare was "greater than".
func setGreater() {
	emitln("CMOVG  EAX, [TRUE]")
	emitln("CMOVLE EAX, [FALSE]")
}

// setLessOrEqual sets primary register if compare was "less or equal".
func setLessOrEqual() {
	emitln("CMOVLE EAX, [TRUE]")
	emitln("CMOVG  EAX, [FALSE]")
}

// setGreaterOrEqual sets primary register if compare was "greater or equal".
func setGreaterOrEqual() {
	emitln("CMOVGE EAX, [TRUE]")
	emitln("CMOVL  EAX, [FALSE]")
}

// testParity tests primary register for parity (even/odd).
func testParity() {
	emitln("TEST EAX, 1")
}

// setOdd sets primary register if parity test was "odd".
func setOdd() {
	emitln("CMOVPO EAX, [TRUE]")
	emitln("CMOVPE EAX, [FALSE]")
}

// branch jumps unconditional.
func branch(L string) {
	emitln("JMP " + L)
}

// branchFalse branches if primary register is false.
func branchFalse(L string) {
	emitln("TEST EAX, -1") // -1 is true
	emitln("JE " + L)
}

// inputNumber reads a number into the primary register.
func inputNumber() {
	emitln("CALL SCANN")
}

// printNumber prints primary register followed by a newline.
func printNumber() {
	emitln("CALL PRINTN")
	emitln("CALL NEWLINE")
}
