package main

import (
	"fmt"
	"errors"
)


type TokenKind int

const (
	TOKEN_IDENTIFIER TokenKind = iota
	TOKEN_DOT
	TOKEN_TERMINAL
	TOKEN_UNKNOWN_TERMINAL
	TOKEN_UNEXPECTED_OPERATOR_PLACE

	TOKEN_OP_OR
	TOKEN_OP_AND

	TOKEN_REPEAT_STAR
	TOKEN_REPEAT_STAR_ONCE
	TOKEN_REPEAT_PLUS
	TOKEN_REPEAT_PLUS_ONCE
	TOKEN_REPEAT_QUESTION

	TOKEN_OPEN_PARENTHESE
	TOKEN_CLOSE_PARENTHESE

	TOKEN_GROUP
	TOKEN_GROUP_OPENER
	TOKEN_GROUP_CLOSER
	TOKEN_GROUP_PARENTHESE
)

var (
	ERROR_MATCHING_OPERATOR_NOT_FOUND error = errors.New("Matching operator not found")
	ERROR_UNARY_OPERATOR_FOUND_AT_UNEXPECTED_PLACE error = errors.New("Unary operator found at unexpected place")
	ERROR_UNEXPECTED_CHARACTER_FOR_LEXER error = errors.New("Lexer was not able to identify a character")
	ERROR_UNEXPECTED_TOKEN_FOR_PARSER error = errors.New("Parser was not able to identify the token(s)")
	ERROR_NOT_FOUND_GROUP_CLOSER_OPERATOR_PARSER error = errors.New("Group closer not found while parsing")
	ERROR_EXPECTED_TERMINAL_TOKEN_FOR_PARSER error = errors.New("Expected terminal token while parsing, got something else")
)

type RegexMatch struct {
	Start	int
	End	int
}

type LexerError struct {
}

type ParserError struct {
	Err	error 
	Node	*NodeTree
}

func (e ParserError) Error() string {
	str := fmt.Sprintf("Error on Node : %s", e.Node)
	str += fmt.Sprintf("\n message: %s", e.Err.Error())

	return str
}

type Token struct {
	Kind	TokenKind
	Value	[]byte
}

func (t Token) String() string {
	str := fmt.Sprintf("{\"Kind\": %d, \"Value\": %q}", int(t.Kind), t.Value)

	return str
}

type NodeTree struct {
	Kind	TokenKind
	Value	Token
	Childreen	[]NodeTree
}

func (n NodeTree) String() string {
	if n.Childreen == nil {
		str := fmt.Sprintf("{\"Kind\": %d, \"Value\": %s, \"Childreen\": %s}", n.Kind, n.Value,  "[]")

		return str
	}

	arr := "["
	for _, child := range n.Childreen {
		// str = fmt.Sprintf("{\"Kind\": %d, \"Value\": %s, \"Childreen\": %s}", n.Kind, n.Value, n.Childreen)
		arr += fmt.Sprintf("%s, ", child)
	}

	arr = arr[:len(arr) - 2]
	arr += "]"

	str := fmt.Sprintf("{\"Kind\": %d, \"Value\": %s, \"Childreen\": %s}", n.Kind, n.Value, arr)

	return str
}

