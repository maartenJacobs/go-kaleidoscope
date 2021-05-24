package main

/*
#include "llvm-c/Core.h"
*/
import "C"

func main() {
	// Create the execution context.
	var context C.LLVMContextRef = C.LLVMContextCreate()
	defer C.LLVMContextDispose(context)

	// Create a module to hold the IR code.
	var mod C.LLVMModuleRef = C.LLVMModuleCreateWithNameInContext(C.CString("my module"), context)
	defer C.LLVMDisposeModule(mod)
}
