package main

/*
#include "llvm-c/Core.h"
*/
import "C"
import (
	"fmt"
	"os"

	gokaleidoscope "github.com/maartenjacobs/go-kaleidoscope/lib"
)

func main() {
	var lex gokaleidoscope.Lexer
	lex.Init(os.Stdin)
	for token := lex.Token(); token != gokaleidoscope.TOKEN_EOF; token = lex.Token() {
		switch token {
		case gokaleidoscope.TOKEN_IDENTIFIER:
			fmt.Print("Identifier: ")
			fmt.Print(lex.LastIdent)
			fmt.Println()
		}
	}

	// Create the execution context.
	var context C.LLVMContextRef = C.LLVMContextCreate()
	defer C.LLVMContextDispose(context)

	// Create a module to hold the IR code.
	var mod C.LLVMModuleRef = C.LLVMModuleCreateWithNameInContext(C.CString("my module"), context)
	defer C.LLVMDisposeModule(mod)
}
