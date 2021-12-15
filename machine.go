package main

import (
	"fmt"
)

type Value interface{}

type Param string

type Machine interface {
	Push(v Value)
	Pop() Value
	Call(Thunk)
	Lookup(sym string) Value
	Bind(sym string, v Value)
}

type Thunk interface {
	Call(Machine)
}

type Word struct {
	Name string
	F    func(Machine)
}

func (w Word) Call(m Machine) {
	w.F(m)
}

type Dict map[string]Value // XXX need to be immutable?

func NewInterpreter() *Interpreter {
	return &Interpreter{
		dict: make(map[string]Value),
	}
}

type Interpreter struct {
	stack []Value
	dict  map[string]Value
}

func (interp *Interpreter) Push(v Value) {
	interp.stack = append(interp.stack, v)
}

func (interp *Interpreter) Pop() Value {
	n := len(interp.stack) - 1
	v := interp.stack[n]
	interp.stack = interp.stack[:n]
	return v
}

func (interp *Interpreter) Bind(sym string, v Value) {
	interp.dict[sym] = v
}

func (interp *Interpreter) Lookup(sym string) Value {
	v, ok := interp.dict[sym]
	if !ok {
		panic(fmt.Errorf("unbound: %q", sym))
	}
	return v
}

func (interp *Interpreter) Call(t Thunk) {
	t.Call(interp)
}

var Drop = Word{
	Name: "drop",
	F: func(m Machine) {
		_ = m.Pop()
	},
}

type Push struct {
	Value Value
}

func (push Push) Call(m Machine) {
	m.Push(push.Value)
}

type Lookup struct {
	Sym string
}

func (lookup Lookup) Call(m Machine) {
	m.Push(m.Lookup(lookup.Sym))
}

var Print = Word{
	Name: "print",
	F: func(m Machine) {
		v := m.Pop()
		fmt.Println(v)
	},
}

var Add = Word{
	Name: "add",
	F: func(m Machine) {
		x := m.Pop().(int)
		y := m.Pop().(int)
		m.Push(x + y)
	},
}

var Let = Word{
	Name: "let",
	F: func(m Machine) {
		x := m.Pop()
		sym := m.Pop().(string)
		m.Bind(sym, x)
		m.Push(x)
	},
}

type Compiler struct {
	parent Machine
	stack  []Value
	dict   map[string]Value
	quote  Quote
	depth  int
}

func NewCompiler(parent Machine) *Compiler {
	return &Compiler{
		parent: parent,
		dict:   make(map[string]Value),
	}
}

type Quote []Thunk

func (q Quote) Call(m Machine) {
	for _, t := range q {
		t.Call(m)
	}
}

func (comp *Compiler) Push(v Value) {
	comp.suffix(Push{v})
	comp.stack = append(comp.stack, v)
}

func (comp *Compiler) Pop() Value {
	n := len(comp.stack) - 1
	if n < 0 {
		return Unknown{}
	}
	v := comp.stack[n]
	comp.stack = comp.stack[:n]
	return v
}

func (comp *Compiler) Bind(sym string, v Value) {
	comp.dict[sym] = v
}

func (comp *Compiler) Lookup(sym string) Value {
	if v, ok := comp.dict[sym]; ok {
		return v
	}
	return comp.parent.Lookup(sym)
}

func (comp *Compiler) Call(t Thunk) {
	comp.suffix(t)
	comp.depth++
	LiftThunk(t).Call(comp)
	comp.depth--
}

func (comp *Compiler) suffix(t Thunk) {
	if comp.depth == 0 {
		if push, ok := t.(Push); ok && push.Value == (Unknown{}) {
			panic("SUFFIXING unknown")
		}
		comp.quote = append(comp.quote, t)
	}
}

var Swap = Word{
	Name: "swap",
	F: func(m Machine) {
		x := m.Pop()
		y := m.Pop()
		m.Push(x)
		m.Push(y)
	},
}

type Unknown struct{}

func LiftThunk(t Thunk) Thunk {
	switch t := t.(type) {
	case Push, Lookup:
		return t
	case Word:
		switch t.Name {
		case "add":
			return Word{
				Name: "lift{add}",
				F: func(m Machine) {
					m.Pop()
					m.Pop()
					m.Push(Unknown{})
				},
			}
		case "let":
			return Let
		case "swap":
			return Swap
		default:
		}
	}
	panic(fmt.Errorf("cannot lift thunk: %#v", t))
}
