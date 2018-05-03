package compiler

import (
	"fmt"
	"io"
	"strconv"
)

var (
	out     io.Writer // Output stream
	labelno int       // Label Counter
)

// write writes to the output stream.
func write(a ...interface{}) {
	fmt.Fprint(out, a...)
}

// writeln writes to the output stream, followed by a newline.
func writeln(a ...interface{}) {
	write(a...)
	fmt.Fprintln(out)
}

func initCode(w io.Writer) {
	out = w
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
