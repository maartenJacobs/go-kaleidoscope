# Kaleidscope LLVM tutorial with Go

This is an adaptation of the [LLVM Kaleidscope tutorial](https://llvm.org/docs/tutorial/MyFirstLanguageFrontend/index.html)
for Go. The C bindings are used, through [cgo](https://golang.org/cmd/cgo/), instead of the C++ library.

## Goals

Same as the LLVM Kaleidscope tutorial, this adaptation builds a compiler for a simple language called Kaleidscope,
e.g.

```
# Compute the x'th fibonacci number.
def fib(x)
  if x < 3 then
    1
  else
    fib(x-1)+fib(x-2)
```

## Build instructions

1. Install the LLVM library, including the headers. I've only used LLVM 10 so far.
2. Run `CGO_CFLAGS="-I/usr/lib/llvm-10/include" CGO_LDFLAGS="-lLLVM-10" go build lib/cmd/main.go`, substituting
"/usr/lib/llvm-10/include" and "LLVM-10" for your LLVM version.

## Caveats and alterations

Besides the obvious not-C++ but Go, this adaptation:

* Uses the Visitor pattern for the code generator to decouple the parser and the code generator.
* Does not have a working REPL. I usually test with files rather than an interactive REPL so I feel it's not
a high priority.
* Uses the Go standard library package "text/scanner" to implement the lexer. Life is too short to write lexers
by hand.
