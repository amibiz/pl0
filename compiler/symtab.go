package compiler

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"
)

type class int

const (
	headCls class = iota
	varCls
	constCls
	procCls
)

func (c class) String() string {
	switch c {
	case varCls:
		return "VAR"
	case constCls:
		return "CONST"
	case procCls:
		return "PROCEDURE"
	default:
		return "kind(" + strconv.Itoa(int(c)) + ")"
	}
}

type object struct {
	name string
	kind class
	lev  int
	next *object
	dsc  *object
	val  string
	pos  int
}

var universe, topScope *object

func openScope() {
	topScope = &object{kind: headCls, dsc: topScope, next: nil}
}

func closeScope() {
	topScope = topScope.dsc
}

func init() {
	topScope = nil
	openScope()
	universe = topScope
}

// newObj creates a new object and places it in the symbol table.
func newObj(id string, class class) *object {
	var obj, x *object
	x = topScope
	for x.next != nil && x.next.name != id {
		x = x.next
	}
	if x.next == nil {
		obj = new(object)
		obj.name = id
		obj.kind = class
		obj.next = nil
		x.next = obj
	} else {
		duplicate(id)
	}
	return obj
}

// find traverses the symbol table looking for a matching object.
func find(id string) *object {
	var s, x *object
	s = topScope
	for {
		x = s.next
		for x != nil && x.name != id {
			x = x.next
		}
		s = s.dsc
		if x != nil || s == nil {
			break
		}
	}
	if x == nil {
		undefined(id)
	}
	return x
}

// DumpTable dumps the symbol table.
func DumpTable() error {
	const padding = 3
	w := tabwriter.NewWriter(os.Stderr, 0, 0, padding, ' ', 0)
	fmt.Fprintln(w, "Symbol\tClass\tValue\tLevel\tPosition\t")
	fmt.Fprintln(w, "------\t-----\t-----\t-----\t--------\t")
	fmt.Fprintln(w, "\t\t\t\t\t")
	dump(w, universe)
	return w.Flush()
}

func dump(w io.Writer, x *object) {
	if x == nil {
		return
	}
	for {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t\n",
			x.name, x.kind, x.val, x.lev, x.pos)
		dump(w, x.dsc)
		x = x.next
		if x == nil {
			break
		}
	}
}
