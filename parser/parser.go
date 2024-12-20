package parser

import (
	"fmt"
	"strconv"
	"thecarrionlang/ast"
	"thecarrionlang/lexer"
	"thecarrionlang/token"
)

const (
	LOWEST      int = iota
	ASSIGN          // =
	LOGICAL_OR      // or
	LOGICAL_AND     // and
	EQUALS          // ==, !=
	LESSGREATER     // >, <
	SUM             // +, -
	PRODUCT         // *, /
	PREFIX          // -X, !X, ++X, --X
	CALL            // myFunction(X)
	POSTFIX         // X++, X--
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:          ASSIGN,
	token.EQ:              EQUALS,
	token.NOT_EQ:          EQUALS,
	token.LT:              LESSGREATER,
	token.GT:              LESSGREATER,
	token.PLUS:            SUM,
	token.MINUS:           SUM,
	token.SLASH:           PRODUCT,
	token.ASTERISK:        PRODUCT,
	token.PLUS_INCREMENT:  POSTFIX,
	token.MINUS_DECREMENT: POSTFIX,
	token.LPAREN:          CALL,
	token.OR:              LOGICAL_OR,
	token.AND:             LOGICAL_AND,
}

type (
	prefixParseFn  func() ast.Expression
	infixParseFn   func(ast.Expression) ast.Expression
	postfixParseFn func(ast.Expression) ast.Expression
)

type Parser struct {
	l         *lexer.Lexer
	currToken token.Token
	peekToken token.Token
	errors    []string

	prefixParseFns    map[token.TokenType]prefixParseFn
	infixParseFns     map[token.TokenType]infixParseFn
	postfixParseFns   map[token.TokenType]postfixParseFn
	statementParseFns map[token.TokenType]func() ast.Statement
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.postfixParseFns = make(map[token.TokenType]postfixParseFn)
	p.statementParseFns = make(map[token.TokenType]func() ast.Statement)

	// Register prefix parsers
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.PLUS_INCREMENT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS_DECREMENT, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	// Register infix parsers
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)

	// Register postfix parsers
	p.registerPostfix(token.PLUS_INCREMENT, p.parsePostfixExpression)
	p.registerPostfix(token.MINUS_DECREMENT, p.parsePostfixExpression)

	// Register statement parsers
	p.registerStatement(token.RETURN, p.parseReturnStatement)
	p.registerStatement(token.IF, p.parseIfStatement)
	p.registerStatement(token.FOR, p.parseForStatement)
	p.registerStatement(token.SPELL, p.parseFunctionDefinition)

	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) registerPostfix(tokenType token.TokenType, fn postfixParseFn) {
	p.postfixParseFns[tokenType] = fn
}

func (p *Parser) registerStatement(tokenType token.TokenType, fn func() ast.Statement) {
	p.statementParseFns[tokenType] = fn
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.currToken, Value: p.currTokenIs(token.TRUE)}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.currToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	if fn, ok := p.statementParseFns[p.currToken.Type]; ok {
		return fn()
	}

	// Handle assignment statements
	if p.currToken.Type == token.IDENT && p.peekTokenIs(token.ASSIGN) {
		return p.parseAssignmentStatement()
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseAssignmentStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{Token: p.currToken}

	stmt.Name = &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Literal,
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.currToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) currTokenIs(t token.TokenType) bool {
	return p.currToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		fmt.Printf("expectPeek: expected %s, and got %s\n", t, p.peekToken.Type)
		return false
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.currToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.NEWLINE) && precedence < p.peekPrecedence() {
		// Handle postfix operators
		if postfixFn, ok := p.postfixParseFns[p.peekToken.Type]; ok {
			p.nextToken()
			leftExp = postfixFn(leftExp)
			continue
		}

		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) currPrecedence() int {
	if p, ok := precedences[p.currToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Literal,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.currToken}

	value, err := strconv.ParseInt(p.currToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
		Left:     left,
	}

	precedence := p.currPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.PostfixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
		Left:     left,
	}

	return expression
}

func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{Token: p.currToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken() // Move to the token after COLON

	if p.currTokenIs(token.NEWLINE) {
		// Multiline block
		if !p.expectPeek(token.INDENT) {
			return nil
		}
		stmt.Consequence = p.parseBlockStatement()
	} else {
		// Inline statement
		stmt.Consequence = &ast.BlockStatement{
			Statements: []ast.Statement{p.parseStatement()},
		}
	}

	// Handle optional else
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken() // Move to the token after COLON

		if p.currTokenIs(token.NEWLINE) {
			// Multiline block
			if !p.expectPeek(token.INDENT) {
				return nil
			}
			stmt.Alternative = p.parseBlockStatement()
		} else {
			// Inline statement
			stmt.Alternative = &ast.BlockStatement{
				Statements: []ast.Statement{p.parseStatement()},
			}
		}
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currToken}
	block.Statements = []ast.Statement{}

	// Consume the INDENT token
	if p.currTokenIs(token.INDENT) {
		p.nextToken()
	}

	for !p.currTokenIs(token.DEDENT) && !p.currTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{
		Token:    p.currToken,
		Function: function,
	}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		p.peekError(token.RPAREN)
		return nil
	}

	return args
}

func (p *Parser) parseForStatement() ast.Statement {
	stmt := &ast.ForStatement{Token: p.currToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Variable = &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Literal,
	}

	if !p.expectPeek(token.IN) {
		return nil
	}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	if !p.expectPeek(token.NEWLINE) {
		return nil
	}

	if !p.expectPeek(token.INDENT) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.COLON) {
			return nil
		}

		if !p.expectPeek(token.NEWLINE) {
			return nil
		}

		if !p.expectPeek(token.INDENT) {
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseFunctionDefinition() ast.Statement {
	stmt := &ast.FunctionDefinition{Token: p.currToken}

	// Expect the function name
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Literal,
	}

	// Expect the parameter list
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.COLON) {
		return nil
	}

	// Check the next token after COLON
	p.nextToken()

	if p.currTokenIs(token.NEWLINE) {
		// We expect an INDENT if it's a multiline function
		if !p.expectPeek(token.INDENT) {
			return nil
		}
		stmt.Body = p.parseBlockStatement()
	} else {
		// Inline statement (function with a single statement)
		stmt.Body = &ast.BlockStatement{
			Statements: []ast.Statement{p.parseStatement()},
		}
	}

	return stmt
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Literal,
	}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{
			Token: p.currToken,
			Value: p.currToken.Literal,
		}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}
