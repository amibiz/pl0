package compiler

import (
	"bufio"
	"io"
)

const eot = 0x4 // End-of-Transmission (Ctrl+D / ^D)

var (
	in     *bufio.Reader // Input stream
	look   byte          // Lookahead character
	token  Token         // Encoded token
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
	for look != '}' {
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
	// Check keywords table
	if tok, ok := keywords[text]; ok {
		token = tok
	} else {
		token = IDENT
	}
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
	token = NUMBER
}

var singles = [256]Token{
	'.': PERIOD,
	',': COMMA,
	';': SEMICOLON,
	'!': SEND,
	'?': RECV,
	'(': LPAREN,
	')': RPARAN,

	'=': EQL,
	'#': NEQ,

	'*': TIMES,
	'/': DIV,
	'+': PLUS,
	'-': MINUS,
}

// next scans the input stream for the next token.
func next() {
	skipWhite()
	if look == eot {
		token, text = EOF, tokens[EOF]
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
	if tok := singles[look]; tok != NULL {
		token, text = tok, tokens[token]
		getChar()
		return
	}
	switch look {
	case ':':
		token, text = follow('=', BECOMES, NULL)
	case '>':
		token, text = follow('=', GEQ, GRT)
	case '<':
		token, text = follow('=', LEQ, LSS)
	default:
		report("illegal character '" + string(look) + "'")
	}
}

func follow(expect byte, fyes Token, fno Token) (Token, string) {
	getChar()
	if look == expect {
		getChar()
		return fyes, tokens[fyes]
	}
	return fno, tokens[fno]
}
