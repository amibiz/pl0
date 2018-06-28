package token

type Token int

const (
	NULL Token = iota
	EOF

	PERIOD    // .
	COMMA     // ,
	SEMICOLON // ;
	BECOMES   // :=
	RECV      // ?
	SEND      // !
	LPAREN    // (
	RPARAN    // )

	relop_start
	EQL // =
	NEQ // #
	LSS // <
	LEQ // <=
	GRT // >
	GEQ // >=
	relop_end

	addop_start
	PLUS  // +
	MINUS // -
	addop_end

	mulop_start
	TIMES // *
	DIV   // /
	mulop_end

	IDENT
	NUMBER

	// Keywords
	keywords_start
	CONST
	VAR
	PROCEDURE
	CALL
	BEGIN
	END
	IF
	THEN
	WHILE
	DO
	ODD
	keywords_end
)

var tokens = [...]string{
	NULL: "NULL",
	EOF:  "EOF",

	PERIOD:    ".",
	COMMA:     ",",
	SEMICOLON: ";",
	BECOMES:   ":=",
	RECV:      "?",
	SEND:      "!",
	LPAREN:    "(",
	RPARAN:    ")",

	EQL: "=",
	NEQ: "#",
	LSS: "<",
	LEQ: "<=",
	GRT: ">",
	GEQ: ">=",

	TIMES: "*",
	DIV:   "/",
	PLUS:  "+",
	MINUS: "-",

	IDENT:  "IDENT",
	NUMBER: "NUMBER",

	CONST:     "CONST",
	VAR:       "VAR",
	PROCEDURE: "PROCEDURE",
	CALL:      "CALL",
	BEGIN:     "BEGIN",
	END:       "END",
	IF:        "IF",
	THEN:      "THEN",
	WHILE:     "WHILE",
	DO:        "DO",
	ODD:       "ODD",
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keywords_start + 1; i < keywords_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup identifies a keyword or IDENT (if not a keyword).
func Lookup(text string) Token {
	if tok, ok := keywords[text]; ok {
		return tok
	}
	return IDENT
}

func (tok Token) String() string {
	return tokens[int(tok)]
}

// IsAddop identifies an addition operator.
func (tok Token) IsAddop() bool {
	return addop_start < tok && tok < addop_end
}

// IsMulop identifies a multiplication operator.
func (tok Token) IsMulop() bool {
	return mulop_start < tok && tok < mulop_end
}

// IsRelop identifies a relation operator.
func (tok Token) IsRelop() bool {
	return relop_start < tok && tok < relop_end
}
