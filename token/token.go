// token/token.go
package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"
	NEWLINE TokenType = "NEWLINE"
	INDENT  TokenType = "INDENT"
	DEDENT  TokenType = "DEDENT"

	// Identifiers and Literals
	IDENT TokenType = "IDENT"
	INT   TokenType = "INT"
	FLOAT TokenType = "FLOAT"

	// Operators
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"
	MOD      TokenType = "%"
	EQ       TokenType = "=="
	NOT_EQ   TokenType = "!="
	LT       TokenType = "<"
	GT       TokenType = ">"
	LE       TokenType = "<="
	GE       TokenType = ">="

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"
	PIPE      TokenType = "|"
	DOT       TokenType = "."

	LPAREN TokenType = "("
	RPAREN TokenType = ")"
	LBRACE TokenType = "{"
	RBRACE TokenType = "}"
	LBRACK TokenType = "["
	RBRACK TokenType = "]"

	// Keywords
	VAR       TokenType = "VAR"
	SPELL     TokenType = "SPELL"
	SPELLBOOK TokenType = "SPELLBOOK"
	TRUE      TokenType = "TRUE"
	FALSE     TokenType = "FALSE"
	IF        TokenType = "IF"
	ELIF      TokenType = "ELIF"
	ELSE      TokenType = "ELSE"
	FOR       TokenType = "FOR"
	IN        TokenType = "IN"
	WHILE     TokenType = "WHILE"
	STOP      TokenType = "STOP"
	SKIP      TokenType = "SKIP"
	IGNORE    TokenType = "IGNORE"
	RETURN    TokenType = "RETURN"

	// Logical Operators
	AND TokenType = "AND"
	OR  TokenType = "OR"
	NOT TokenType = "NOT"
)

var keywords = map[string]TokenType{
	"var":       VAR,
	"spell":     SPELL,
	"spellbook": SPELLBOOK,
	"true":      TRUE,
	"false":     FALSE,
	"if":        IF,
	"elif":      ELIF,
	"else":      ELSE,
	"for":       FOR,
	"in":        IN,
	"while":     WHILE,
	"stop":      STOP,
	"skip":      SKIP,
	"ignore":    IGNORE,
	"and":       AND,
	"or":        OR,
	"not":       NOT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// LookupIndent determines the TokenType based on the indentation string.
func LookupIndent(indent string) TokenType {
	indentLevels := map[int]TokenType{
		0: DEDENT,
		4: INDENT, // 4 spaces
		8: INDENT, // 8 spaces, etc.
		// Add more levels as needed
	}

	length := len(indent)
	if tok, ok := indentLevels[length]; ok {
		return tok
	}
	return ILLEGAL
}

