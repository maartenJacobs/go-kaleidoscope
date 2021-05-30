package main

import (
	"fmt"
	"os"

	gokaleidoscope "github.com/maartenjacobs/go-kaleidoscope/lib"
)

func handleDefinition(parser *gokaleidoscope.Parser) {
	_, err := parser.ParseDefinition()
	if err != nil {
		// Skip token for error recovery.
		parser.Next()
		fmt.Println(err)
		fmt.Println("Unable to parse definition")
	} else {
		fmt.Println("Parsed definition")
	}
}

func handleExtern(parser *gokaleidoscope.Parser) {
	_, err := parser.ParseExtern()
	if err != nil {
		// Skip token for error recovery.
		parser.Next()
		fmt.Println(err)
		fmt.Println("Unable to parse external")
	} else {
		fmt.Println("Parsed external")
	}
}

func handleTopLevelExpr(parser *gokaleidoscope.Parser) {
	_, err := parser.ParseTopLevelExpr()
	if err != nil {
		// Skip token for error recovery.
		parser.Next()
		fmt.Println(err)
		fmt.Println("Unable to parse top-level expression")
	} else {
		fmt.Println("Parsed top-level expression")
	}
}

/// top ::= definition | external | expression | ';'
func mainLoop(parser *gokaleidoscope.Parser) {
	for {
		switch parser.Current {
		case gokaleidoscope.TOKEN_EOF:
			return
		case gokaleidoscope.TOKEN_DEF:
			handleDefinition(parser)
		case gokaleidoscope.TOKEN_EXTERN:
			handleExtern(parser)
		case gokaleidoscope.TOKEN_UNKNOWN:
			if parser.LastChar() == ';' {
				parser.Next() // ignore top-level semicolons.
			} else {
				handleTopLevelExpr(parser)
			}
		default:
			handleTopLevelExpr(parser)
		}
		fmt.Print("ready> ")
	}
}

func main() {
	var parser gokaleidoscope.Parser
	parser.Init(os.Stdin)
	parser.BinOpPrecedence['<'] = 10
	parser.BinOpPrecedence['+'] = 20
	parser.BinOpPrecedence['-'] = 20
	parser.BinOpPrecedence['*'] = 40

	fmt.Print("ready> ")
	parser.Next() // Prime the first token.

	mainLoop(&parser)
}
