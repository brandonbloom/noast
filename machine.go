package main

import "fmt"

type Value = int

type Machine interface {
	Push(v Value)
	Pop() Value
	Call(Quote)
}

type Quote func(Machine)

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

type Interpreter struct {
	stack []Value
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

func (interp *Interpreter) Call(q Quote) {
	q(interp)
}

func Print(m Machine) {
	v := m.Pop()
	fmt.Println(v)
}

func Add(m Machine) {
	x := m.Pop()
	y := m.Pop()
	m.Push(x + y)
}
