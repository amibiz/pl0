package compiler

import (
	"bufio"
	"io"

	"pl0/compiler/token"
)

const eot = 0x4 // End-of-Transmission (Ctrl+D / ^D)

var (
	in     *bufio.Reader // Input stream
	look   byte          // Lookahead character
	tok    token.Token   // Encoded token
	text   string        // Unencoded token
	lineno int           // Current lineno number
)

func initScanner(r io.Reader) {
	in = bufio.NewReader(r)
	getChar()
	lineno = 1
}

// getChar reads new character from the input stream.
func getChar() {
	if b, err := in.ReadByte(); err == io.EOF {
		look = eot
	} else {
		look = b
	}
	if look == '\n' {
		lineno++
	}
}

// isAlpha recognizes an alpha character.
func isAlpha(c byte) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z')
}

// isDigit recognizes a decimal digit.
func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// isAlNum recognizes an Alphanumeric Character.
func isAlNum(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

// isWhite recognizes white space.
func isWhite(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '{':
		return true
	}
	return false
}

// skipWhite skips over leading white space or comment field.
func skipWhite() {
	for isWhite(look) {
		for look == '{' {
			skipComment()
		}
		getChar()
	}
}

// skipComment skips a comment field.
func skipComment() {
	for look != '}' && look != eot {
		getChar()
		if look == '{' {
			skipComment()
		}
	}
	getChar()
}

// scanIdent scans an identifier.
func scanIdent() {
	text = ""
	if !isAlpha(look) {
		expected("identifier", string(look))
	}
	for isAlNum(look) {
		text += string(look)
		getChar()
	}
	tok = token.Lookup(text)
}

// scanNumber scans a Number.
func scanNumber() {
	text = ""
	if !isDigit(look) {
		expected("number", string(look))
	}
	for isDigit(look) {
		text += string(look)
		getChar()
	}
	tok = token.NUMBER
}

var singles = [256]token.Token{
	'.': token.PERIOD,
	',': token.COMMA,
	';': token.SEMICOLON,
	'!': token.SEND,
	'?': token.RECV,
	'(': token.LPAREN,
	')': token.RPARAN,

	'=': token.EQL,
	'#': token.NEQ,

	'*': token.TIMES,
	'/': token.DIV,
	'+': token.PLUS,
	'-': token.MINUS,
}

// next scans the input stream for the next token.
func next() {
	skipWhite()
	if look == eot {
		tok, text = token.EOF, token.EOF.String()
		return
	}
	if isAlpha(look) {
		scanIdent()
		return
	}
	if isDigit(look) {
		scanNumber()
		return
	}
	if t := singles[look]; t != token.NULL {
		tok, text = t, t.String()
		getChar()
		return
	}
	switch look {
	case ':':
		tok, text = follow('=', token.BECOMES, token.NULL)
	case '>':
		tok, text = follow('=', token.GEQ, token.GRT)
	case '<':
		tok, text = follow('=', token.LEQ, token.LSS)
	default:
		report("illegal character '" + string(look) + "'")
	}
}

func follow(expect byte, fyes, fno token.Token) (token.Token, string) {
	getChar()
	if look == expect {
		getChar()
		return fyes, fyes.String()
	}
	return fno, fno.String()
}
