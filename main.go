package main

import (
	"context"
	"fmt"
	"strings"
)

type SourceContext interface {
	context.Context
	Source
	Machine
}

type Source interface {
	PeekChar() rune
	NextChar()
	SourceIndex() int
}

func main() {
	//	code := `
	//(let x 2)
	//(add x 3)
	//`
	code := `
(def inc (x)
  (add x 1))
(inc 5)
`
	//	code := `
	//(def (fact x)
	//  (if (= x 0)
	//    1
	//    (* x (fact (- x 1)))))
	//(fact 5)
	//`

	ctx := struct {
		context.Context
		Source
		Machine
	}{
		Context: context.Background(),
		Source:  NewStringSource(code),
		Machine: NewInterpreter(),
	}
	topLevel(ctx)
}

type StringSource struct {
	s string
	i int
}

func NewStringSource(s string) *StringSource {
	return &StringSource{
		s: s,
	}
}

func (src *StringSource) PeekChar() rune {
	if len(src.s) <= src.i {
		return 0
	}
	return rune(src.s[src.i]) // XXX unicode, etc.
}

func (src *StringSource) NextChar() {
	src.i++
}

func (src *StringSource) SourceIndex() int {
	return src.i
}

func topLevel(ctx SourceContext) {
	for acceptExpr(ctx) {
		ctx.Call(Print)
	}
}

func acceptExprs(ctx SourceContext) {
	for acceptExpr(ctx) {
	}
}

func expectExpr(ctx SourceContext) {
	if !acceptExpr(ctx) {
		panic(fmt.Errorf("expected expr"))
	}
}

func acceptExpr(ctx SourceContext) bool {
	acceptWhite(ctx)
	c := ctx.PeekChar()
	switch {
	case c == 0, c == ')':
		return false
	case isDigit(c):
		ctx.Push(expectNumber(ctx))
	case isSymbolChar(c):
		ctx.Call(Lookup{expectSymbol(ctx)})
	case c == '(':
		expectCall(ctx)
	default:
		panic(fmt.Errorf("unexpected: %q", c))
	}
	return true
}

func acceptWhite(ctx SourceContext) {
	for isWhite(ctx.PeekChar()) {
		ctx.NextChar()
	}
}

func isWhite(c rune) bool {
	return c == ' ' || c == '\n'
}

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

func expectChar(ctx SourceContext, c rune) {
	acceptWhite(ctx)
	x := ctx.PeekChar()
	if x != c {
		panic(fmt.Errorf("expected %q, got %q", c, x))
	}
	ctx.NextChar()
}

func expectNumber(ctx SourceContext) int {
	n := expectDigit(ctx)
	for {
		d, ok := acceptDigit(ctx)
		if !ok {
			break
		}
		n *= 10
		n += d
	}
	return n
}

func expectDigit(ctx SourceContext) int {
	d, ok := acceptDigit(ctx)
	if !ok {
		panic(fmt.Errorf("expected digit"))
	}
	return d
}

func acceptDigit(ctx SourceContext) (d int, ok bool) {
	c := ctx.PeekChar()
	if isDigit(c) {
		ctx.NextChar()
		return int(c) - '0', true
	}
	return 0, false
}

func expectCall(ctx SourceContext) {
	expectList(ctx, func() {
		name := expectSymbol(ctx)
		acceptWhite(ctx)
		switch name {

		case "let":
			ctx.Push(expectSymbol(ctx))
			expectExpr(ctx)
			ctx.Call(Let)

		case "def":
			sym := expectSymbol(ctx)
			compiler := NewCompiler(ctx)
			compileCtx := struct {
				context.Context
				Source
				Machine
			}{
				Context: ctx,
				Source:  ctx,
				Machine: compiler,
			}
			expectParams(compileCtx)
			acceptExpr(compileCtx)
			ctx.Bind(sym, compiler.quote)
			ctx.Push(nil)

		case "add":
			expectExpr(ctx)
			expectExpr(ctx)
			ctx.Call(Add)

		default:
			acceptExprs(ctx)
			ctx.Call(ctx.Lookup(name).(Thunk))
		}
	})
}

func expectList(ctx SourceContext, f func()) {
	expectChar(ctx, '(')
	acceptWhite(ctx)
	f()
	acceptWhite(ctx)
	expectChar(ctx, ')')
}

func expectSymbol(ctx SourceContext) string {
	sym := acceptSymbol(ctx)
	if len(sym) == 0 {
		panic(fmt.Errorf("expected symbol, got %q", ctx.PeekChar()))
	}
	return sym
}

func acceptSymbol(ctx SourceContext) string {
	var sb strings.Builder
	for {
		c := ctx.PeekChar()
		// XXX initial vs continue
		if !isSymbolChar(c) {
			break
		}
		ctx.NextChar()
		sb.WriteRune(c)
	}
	return sb.String()
}

func isSymbolChar(c rune) bool {
	return 'a' <= c && c <= 'z'
}

func expectParams(ctx SourceContext) {
	expectList(ctx, func() {
		var syms []string
		for {
			sym := acceptSymbol(ctx)
			if len(sym) == 0 {
				break
			}
			syms = append(syms, sym)
		}
		for i := len(syms) - 1; i >= 0; i-- {
			ctx.Push(syms[i])
			ctx.Call(Swap)
			ctx.Call(Let)
		}
	})
}
