package main

import "fmt"

type Value interface{}

type Machine interface {
	Push(v Value)
	Pop() Value
	Call(Word)
	Lookup(sym string) Value
	Bind(sym string, v Value)
}

type Word func(Machine)

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

func (interp *Interpreter) Call(q Word) {
	q(interp)
}

func Print(m Machine) {
	v := m.Pop()
	fmt.Println(v)
}

func Add(m Machine) {
	x := m.Pop().(int)
	y := m.Pop().(int)
	m.Push(x + y)
}

func Let(m Machine) {
	x := m.Pop()
	sym := m.Pop().(string)
	m.Bind(sym, x)
	m.Push(x)
}

func Def(m Machine) {
	panic("TODO: Def")
}

type Compiler struct {
	queue Quote
}

//func NewCompiler() *Compiler {
//}

type Quote []Word
