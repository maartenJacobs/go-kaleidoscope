package gokaleidoscope

import (
	"io"
	"strconv"
	"text/scanner"
)

type token uint

const (
	TOKEN_UNKNOWN token = iota
	TOKEN_EOF

	// commands
	TOKEN_DEF
	TOKEN_EXTERN

	// Primary
	TOKEN_IDENTIFIER
	TOKEN_NUMBER
)

type Lexer struct {
	scanner   scanner.Scanner
	LastFloat float64
	LastIdent string
}

func (lex *Lexer) Init(src io.Reader) {
	lex.scanner.Init(src)
	// scanner.ScanFloats includes ints and hex floats.
	// scanner.ScanIdents matches Go identifiers.
	// Both issues are fine for a tutorial though.
	lex.scanner.Mode = scanner.ScanFloats | scanner.ScanIdents
}

func (lex *Lexer) Token() token {
	switch lex.scanner.Scan() {
	case scanner.Float:
		if parsedFloat, parseErr := strconv.ParseFloat(lex.scanner.TokenText(), 64); parseErr != nil {
			// TODO: can scanner scan a float that cannot be parsed by strconv?
			panic("Scanned float could not be parsed by Go")
		} else {
			lex.LastFloat = parsedFloat
		}
		return TOKEN_NUMBER
	case scanner.Ident:
		tokenText := lex.scanner.TokenText()
		if tokenText == "def" {
			return TOKEN_DEF
		} else if tokenText == "extern" {
			return TOKEN_EXTERN
		} else {
			lex.LastIdent = tokenText
			return TOKEN_IDENTIFIER
		}
	case scanner.EOF:
		return TOKEN_EOF
	default:
		return TOKEN_UNKNOWN
	}
}
