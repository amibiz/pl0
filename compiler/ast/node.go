package ast

import (
	"pl0/compiler/token"
)

type Node interface{ node() }

type Stmt interface {
	Node
	stmtNode()
}

type Cond interface {
	Node
	condNode()
}

type Expr interface {
	Node
	exprNode()
}

type Program struct {
	Name string
	Main *Block
}

type Block struct {
	Consts []*ConstDecl
	Vars   []*Ident
	Procs  []*ProcDecl
	Body   Stmt
}

type ConstDecl struct {
	Name  *Ident
	Value *Number
}

type ProcDecl struct {
	Name  *Ident
	Block *Block
}

// Statement nodes.
type (
	AssignStmt struct {
		Lhs *Ident
		Rhs Expr
	}

	CallStmt struct {
		Proc *Ident
	}

	SendStmt struct {
		X Expr
	}

	ReceiveStmt struct {
		Name *Ident
	}

	BeginStmt struct {
		List []Stmt
	}

	IfStmt struct {
		Cond Cond
		Body Stmt
	}

	WhileStmt struct {
		Cond Cond
		Body Stmt
	}
)

// All nodes that implement the Stmt interface
func (*AssignStmt) stmtNode()  {}
func (*CallStmt) stmtNode()    {}
func (*SendStmt) stmtNode()    {}
func (*ReceiveStmt) stmtNode() {}
func (*BeginStmt) stmtNode()   {}
func (*IfStmt) stmtNode()      {}
func (*WhileStmt) stmtNode()   {}

// Condition nodes.
type (
	OddCond struct {
		X Expr
	}

	RelCond struct {
		X  Expr
		Op token.Token
		Y  Expr
	}
)

// All nodes that implement the Cond interface
func (*OddCond) condNode() {}
func (*RelCond) condNode() {}

// Expression nodes.
type (
	Ident struct {
		Name string
	}

	Number struct {
		Value string
	}

	UnaryExpr struct {
		X  Expr
		Op token.Token
	}

	BinaryExpr struct {
		X  Expr
		Op token.Token
		Y  Expr
	}
)

// All nodes that implement the Expr interface
func (*Ident) exprNode()      {}
func (*Number) exprNode()     {}
func (*UnaryExpr) exprNode()  {}
func (*BinaryExpr) exprNode() {}

// All nodes implement the Node interface
func (*Program) node()     {}
func (*Block) node()       {}
func (*ConstDecl) node()   {}
func (*ProcDecl) node()    {}
func (*AssignStmt) node()  {}
func (*CallStmt) node()    {}
func (*SendStmt) node()    {}
func (*ReceiveStmt) node() {}
func (*BeginStmt) node()   {}
func (*IfStmt) node()      {}
func (*WhileStmt) node()   {}
func (*OddCond) node()     {}
func (*RelCond) node()     {}
func (*Ident) node()       {}
func (*Number) node()      {}
func (*UnaryExpr) node()   {}
func (*BinaryExpr) node()  {}
