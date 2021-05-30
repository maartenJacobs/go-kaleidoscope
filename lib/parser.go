package gokaleidoscope

import (
	"errors"
	"io"
)

/*
AST structs.

Includes entry point for visitor pattern.
*/

// Dummy interface for AST nodes.
type ExprAST interface {
	astType() string
	Accept(visitor ExprASTVisitor) error // The error is passed through from the relevant Visit* method.
}

type NumberExprAST struct {
	Val float64
}

func (number *NumberExprAST) astType() string {
	return "NumberExprAST"
}

type VariableExprAST struct {
	Name string
}

func (variable *VariableExprAST) astType() string {
	return "VariableExprAST"
}

type BinaryExprAST struct {
	Op    rune
	Left  ExprAST
	Right ExprAST
}

func (binary *BinaryExprAST) astType() string {
	return "BinaryExprAST"
}

type CallExprAST struct {
	Callee string
	Args   []ExprAST
}

func (call *CallExprAST) astType() string {
	return "CallExprAST"
}

type PrototypeAST struct {
	Callee string
	Args   []string
}

type FunctionAST struct {
	Proto PrototypeAST
	Body  ExprAST
}

/*
AST visitor pattern
*/

type ExprASTVisitor interface {
	VisitNumberExprAST(node *NumberExprAST) error
	VisitVariableExprAST(node *VariableExprAST) error
	VisitBinaryExprAST(node *BinaryExprAST) error
	VisitCallExprAST(node *CallExprAST) error
}

func (number *NumberExprAST) Accept(visitor ExprASTVisitor) error {
	return visitor.VisitNumberExprAST(number)
}

func (binary *BinaryExprAST) Accept(visitor ExprASTVisitor) error {
	return visitor.VisitBinaryExprAST(binary)
}

func (variable *VariableExprAST) Accept(visitor ExprASTVisitor) error {
	return visitor.VisitVariableExprAST(variable)
}

func (call *CallExprAST) Accept(visitor ExprASTVisitor) error {
	return visitor.VisitCallExprAST(call)
}

/*
Parser
*/

type Parser struct {
	Current         token
	BinOpPrecedence map[rune]int
	lexer           *Lexer
}

func (parser *Parser) Init(src io.Reader) {
	parser.lexer = &Lexer{}
	parser.lexer.Init(src)

	parser.BinOpPrecedence = make(map[rune]int)
}

/// Helper function to retrieve the precedence of the current token.
/// If the current token is not an operator, then -1 is returned.
func (parser *Parser) tokenPrecedence() int {
	if parser.Current != TOKEN_UNKNOWN {
		return -1
	}

	if precedence, ok := parser.BinOpPrecedence[parser.lexer.LastChar]; !ok {
		return -1
	} else {
		return precedence
	}
}

/// Match a token without a token class, e.g. '('.
func (parser *Parser) match(char rune) bool {
	return parser.Current == TOKEN_UNKNOWN && parser.lexer.LastChar == char
}

func (parser *Parser) LastChar() rune {
	return parser.lexer.LastChar
}

func (parser *Parser) Next() token {
	parser.Current = parser.lexer.Token()
	return parser.Current
}

/// numberexpr ::= number
func (parser *Parser) ParseNumber() ExprAST {
	expr := NumberExprAST{parser.lexer.LastFloat}
	parser.Next() // Consume the number.
	return &expr
}

/// parenexpr ::= '(' expression ')'
func (parser *Parser) ParseParenExpr() (ExprAST, error) {
	// eat `(`.
	parser.Next()

	val, err := parser.ParseExpr()
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	if !parser.match(')') {
		return nil, errors.New("Expected ')'")
	}

	parser.Next() // Eat ')'
	return val, nil
}

/// identifierexpr
///   ::= identifier
///   ::= identifier '(' expression* ')'
func (parser *Parser) ParseIdentifier() (ExprAST, error) {
	idName := parser.lexer.LastIdent
	parser.Next() // Eat identifier.

	if !parser.match('(') {
		return &VariableExprAST{idName}, nil
	}

	parser.Next() // Eat '('.
	var args []ExprAST
	if !parser.match(')') {
		for {
			arg, err := parser.ParseExpr()
			if err != nil {
				return nil, err
			}
			if arg != nil {
				args = append(args, arg)
			} else {
				return nil, nil
			}

			if parser.match(')') {
				break
			}

			if !parser.match(',') {
				return nil, errors.New("Expected ','")
			}
			parser.Next() // Eat ','.
		}
	}

	parser.Next() // Eat ')'.
	return &CallExprAST{idName, args}, nil
}

/// primary
///   ::= identifierexpr
///   ::= numberexpr
///   ::= parenexpr
func (parser *Parser) ParsePrimary() (ExprAST, error) {
	switch parser.Current {
	case TOKEN_IDENTIFIER:
		return parser.ParseIdentifier()
	case TOKEN_NUMBER:
		return parser.ParseNumber(), nil
	default:
		if parser.match('(') {
			return parser.ParseParenExpr()
		} else {
			return nil, errors.New("Unknown token when expecting an expression")
		}
	}
}

/// binoprhs
///   ::= ('+' primary)*
func (parser *Parser) ParseBinOpRHS(exprPrec int, lhs ExprAST) (ExprAST, error) {
	for {
		// If this is a binop, find its precedence.
		precedence := parser.tokenPrecedence()

		// If this is a binop that binds at least as tightly as the current binop,
		// consume it, otherwise we are done.
		if precedence < exprPrec {
			return lhs, nil
		}

		// Okay, we know this is a binop.
		binOp := parser.lexer.LastChar
		parser.Next() // Eat binary operator.

		// Parse the primary expression after the binary operator.
		rhs, rhsErr := parser.ParsePrimary()
		if rhsErr != nil {
			return nil, rhsErr
		}

		// If BinOp binds less tightly with RHS than the operator after RHS, let
		// the pending operator take RHS as its LHS.
		nextPrecedence := parser.tokenPrecedence()
		if precedence < nextPrecedence {
			var rhsError error
			rhs, rhsError = parser.ParseBinOpRHS(precedence+1, rhs)
			if rhsError != nil {
				return nil, rhsErr
			}
		}

		// Merge LHS/RHS.
		lhs = &BinaryExprAST{binOp, lhs, rhs}
	}
}

/// expression
///   ::= primary binoprhs
///
func (parser *Parser) ParseExpr() (ExprAST, error) {
	left, err := parser.ParsePrimary()
	if err != nil {
		return nil, err
	}

	return parser.ParseBinOpRHS(0, left)
}

/// prototype
///   ::= id '(' id* ')'
func (parser *Parser) ParsePrototype() (*PrototypeAST, error) {
	if parser.Current != TOKEN_IDENTIFIER {
		return nil, errors.New("Expected function name in prototype")
	}

	funcName := parser.lexer.LastIdent
	parser.Next() // Eat function name.

	// Match open braces before arguments.
	if !parser.match('(') {
		return nil, errors.New("Expected '(' in prototype")
	}

	var argNames []string
	for parser.Next() == TOKEN_IDENTIFIER {
		argNames = append(argNames, parser.lexer.LastIdent)
	}

	// Match close braces after arguments.
	if !parser.match(')') {
		return nil, errors.New("Expected ')' in prototype")
	}
	parser.Next() // Eat ')'.

	return &PrototypeAST{funcName, argNames}, nil
}

/// definition ::= 'def' prototype expression
func (parser *Parser) ParseDefinition() (*FunctionAST, error) {
	parser.Next() // Eat 'def'.

	proto, protoErr := parser.ParsePrototype()
	if protoErr != nil {
		return nil, protoErr
	}

	expr, exprErr := parser.ParseExpr()
	if exprErr != nil {
		return nil, exprErr
	}

	return &FunctionAST{*proto, expr}, nil
}

/// toplevelexpr ::= expression
func (parser *Parser) ParseTopLevelExpr() (*FunctionAST, error) {
	expr, exprErr := parser.ParseExpr()
	if exprErr != nil {
		return nil, exprErr
	}

	return &FunctionAST{
		PrototypeAST{Callee: "__anon_expr", Args: make([]string, 0)},
		expr,
	}, nil
}

/// external ::= 'extern' prototype
func (parser *Parser) ParseExtern() (*PrototypeAST, error) {
	parser.Next() // Eat 'extern'.
	return parser.ParsePrototype()
}
