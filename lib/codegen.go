package gokaleidoscope

/*
#include "llvm-c/Core.h"
#include <stdlib.h>

// Helper function to set the arguments array.
void set_arg_in(LLVMValueRef * args, int at, LLVMValueRef arg) {
	args[at] = arg;
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

/// Private visitor that generates LLVM values.
/// This ensures that the Codegen interface remains simple, e.g. Codegen only exposes Generate(ExprAst)
/// instead of VisitNumberExprAST, VisitVariableExprAST, etc.
type astCodegenVisitor struct {
	context     C.LLVMContextRef
	mod         C.LLVMModuleRef
	builder     C.LLVMBuilderRef
	namedValues map[string]C.LLVMValueRef
	LastValue   C.LLVMValueRef // Should be set after calling Visit() and no error is returned.
}

func (visitor *astCodegenVisitor) VisitNumberExprAST(node *NumberExprAST) error {
	visitor.LastValue = C.LLVMConstReal(C.LLVMDoubleTypeInContext(visitor.context), C.double(node.Val))
	return nil
}

func (visitor *astCodegenVisitor) VisitVariableExprAST(node *VariableExprAST) error {
	if value, ok := visitor.namedValues[node.Name]; !ok {
		return errors.New("Unknown variable name")
	} else {
		visitor.LastValue = value
		return nil
	}
}

func (visitor *astCodegenVisitor) VisitBinaryExprAST(node *BinaryExprAST) error {
	leftErr := node.Left.Accept(visitor)
	if leftErr != nil {
		return leftErr
	}
	left := visitor.LastValue

	rightErr := node.Right.Accept(visitor)
	if rightErr != nil {
		return rightErr
	}
	right := visitor.LastValue

	switch node.Op {
	case '+':
		opName := C.CString("addtmp")
		defer C.free(unsafe.Pointer(opName))
		visitor.LastValue = C.LLVMBuildFAdd(visitor.builder, left, right, opName)
		return nil
	case '-':
		opName := C.CString("subtmp")
		defer C.free(unsafe.Pointer(opName))
		visitor.LastValue = C.LLVMBuildFSub(visitor.builder, left, right, opName)
		return nil
	case '*':
		opName := C.CString("multmp")
		defer C.free(unsafe.Pointer(opName))
		visitor.LastValue = C.LLVMBuildFMul(visitor.builder, left, right, opName)
		return nil
	case '<':
		cmpOpName := C.CString("cmptmp")
		defer C.free(unsafe.Pointer(cmpOpName))
		boolOpName := C.CString("booltmp")
		defer C.free(unsafe.Pointer(boolOpName))

		left = C.LLVMBuildFCmp(visitor.builder, C.LLVMRealULT, left, right, cmpOpName)
		// Convert bool 0/1 to double 0.0 or 1.0
		visitor.LastValue = C.LLVMBuildUIToFP(visitor.builder, left, C.LLVMDoubleTypeInContext(visitor.context), boolOpName)
		return nil
	default:
		return errors.New("invalid binary operator")
	}
}

func (visitor *astCodegenVisitor) VisitCallExprAST(node *CallExprAST) error {
	calleeName := C.CString(node.Callee)
	defer C.free(unsafe.Pointer(calleeName))
	callee := C.LLVMGetNamedFunction(visitor.mod, calleeName)
	if callee == nil {
		return errors.New("Unknown function referenced")
	}

	paramCount := C.LLVMCountParamTypes(C.LLVMGetElementType(C.LLVMTypeOf(callee)))
	argCount := C.uint(len(node.Args)) // Not sure why len() returns an int
	if paramCount != argCount {
		return errors.New("Incorrect # arguments passed")
	}

	// Create a C array from the arguments.
	valueSize := C.uint(C.sizeof_LLVMValueRef)
	args := C.malloc(C.ulong(valueSize * argCount))
	defer C.free(unsafe.Pointer(args))
	argNumber := 0
	for arg := range node.Args {
		// Setting the arguments is pretty painful. `unsafe` allows iteration but not setting values?
		C.set_arg_in(&args[0], C.int(argNumber), arg)
		argNumber = argNumber + 1
	}

	opName := C.CString("calltmp")
	defer C.free(unsafe.Pointer(opName))

	return nil
}

/// Generate LLVM values from AST nodes.
type Codegen struct {
	context     C.LLVMContextRef
	mod         C.LLVMModuleRef
	builder     C.LLVMBuilderRef
	namedValues map[string]C.LLVMValueRef
}

/// Initialise the codegen before generation.
func (codegen *Codegen) Init() {
	codegen.namedValues = make(map[string]C.LLVMValueRef)

	// Create the execution context.
	codegen.context = C.LLVMContextCreate()

	// Create the IR builder.
	codegen.builder = C.LLVMCreateBuilderInContext(codegen.context)

	// Create a module to hold the IR code.
	codegen.mod = C.LLVMModuleCreateWithNameInContext(C.CString("my module"), codegen.context)
}

/// Close the resources managed by the code generator.
func (codegen *Codegen) Close() {
	C.LLVMDisposeModule(codegen.mod)
	C.LLVMDisposeBuilder(codegen.builder)
	C.LLVMContextDispose(codegen.context)
}

func (codegen *Codegen) Generate(expr ExprAST) (C.LLVMValueRef, error) {
	visitor := &astCodegenVisitor{
		context:     codegen.context,
		mod:         codegen.mod,
		builder:     codegen.builder,
		namedValues: codegen.namedValues,
		LastValue:   nil,
	}
	if visitErr := expr.Accept(visitor); visitErr != nil {
		return nil, visitErr
	}
	return visitor.LastValue, nil
}
