package main

import (
	// "bufio"
	"bytes"
	"fmt"
	// "os"
)

func main() {
	/*
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		regex := scanner.Text()
		tokens, _ := tokenize([]byte(regex))
		tree, _ := buildParseTree(tokens)

		fmt.Println(tree)
		// fmt.Printf("regex: %q \ntokens : %q\n", regex, tokens)

		fmt.Println()
		for _, token := range tokens {
			fmt.Printf("-- %q ", token.Value)
		}
		fmt.Println()
	}
	*/

	/*
	str := ""
	str = "a|b(c|d)"
	str = "(albion+)+?belfast*?Ricar+(.*?)+.."
	// str = "(albion+?"
	// str = "albion+)dodo()("

	regex := []byte(str)
	tokens, err := tokenize(regex)
	if err != nil {
		fmt.Printf("Tokenization error: %s", err.Error())
	}

	// fmt.Println(tokens)

	tree, errs := buildParseTree(tokens)
	if errs != nil {
		fmt.Println("Error parser: ", errs)
	}

	fmt.Println(tree)
	_ = tree
	*/

	// text := []byte("hello")
	// tree := new(NodeTree)

	var reg []byte = nil
	// reg = []byte("a|b|c|ab|a.")
	reg = []byte("a(a|b|c)?")
	lex, _ := tokenize(reg)
	tree, _ := buildParseTree(lex)

	text := []byte("abaaac")

	matched := evaluateRegex(text, tree, nil)

	var results []string
	for _, match := range matched {
		result := string(text[match.Start:match.End])
		results = append(results, result)
	}

	fmt.Println("final results: ", results)
}

func evaluateRegex(text []byte, tree *NodeTree, data []RegexMatch) []RegexMatch {
	if tree == nil {
		fmt.Println("parent tree == nil")
		return nil
	}

	if len(tree.Childreen) == 2 {
		fmt.Println("childreen == 2")

		switch tree.Kind {
		case TOKEN_OP_AND:
			data = evaluateRegex(text, &tree.Childreen[0], data)
			data = evaluateRegex(text, &tree.Childreen[1], data)
		case TOKEN_OP_OR:
			right := evaluateRegex(text, &tree.Childreen[0], data)
			left := evaluateRegex(text, &tree.Childreen[1], data)
			data = append(right, left...)
		default:
			data = nil
			fmt.Println("Unexpected error while evaluating regex for childreen == 2")
		}

		return data
	}

	if len(tree.Childreen) == 1 {
		fmt.Println("childreen == 1")

		if tree.Kind == TOKEN_OPEN_PARENTHESE {
			data = evaluateRegex(text, &tree.Childreen[0], data)
		} else if tree.Kind == TOKEN_REPEAT_PLUS || tree.Kind == TOKEN_REPEAT_PLUS_ONCE {
			freshData := data
			data = []RegexMatch{}

			for {
				freshData = evaluateRegex(text, &tree.Childreen[0], freshData)
				if len(freshData) == 0 {
					break
				}

				data = append(data, freshData...)
			}
		} else if tree.Kind == TOKEN_REPEAT_STAR || tree.Kind == TOKEN_REPEAT_STAR_ONCE {
			freshData := data

			for {
				freshData = evaluateRegex(text, &tree.Childreen[0], freshData)
				if len(freshData) == 0 {
					break
				}

				data = append(data, freshData...)
			}
		} else if tree.Kind == TOKEN_REPEAT_QUESTION {
			freshData := data

			freshData = evaluateRegex(text, &tree.Childreen[0], freshData)

			data = append(data, freshData...)
		} else {
			data = nil
			fmt.Println("Unexpected error while evaluating regex for childreen == 1")
		}

		return data
	}

	// Terminal Node
	var matched []RegexMatch

	matcher := tree.Value.Value
	if tree.Kind == TOKEN_DOT {
		matcher = nil
	}

	if data == nil {
		matched = matchFullText(text, matcher)
	} else {
		matched = matchPartialText(text, matcher, data)
	}

	fmt.Println("evaluate matched: ", matched)

	return matched
}

// If matcher == nil, match every character
func matchFullText (text []byte, matcher []byte) []RegexMatch {
	matched := []RegexMatch{}

	isDotOperator := matcher == nil

	for index, _ := range text {
		if isDotOperator {
			if index < len(text) {
				newMatch := RegexMatch { Start: index, End: index + 1 }
				matched = append(matched, newMatch)
			}

			continue
		}

		if len(matcher) + index > len(text) {
			break
		}

		counter := 0
		for _, mat := range matcher {
			char := text[index + counter]

			if char == mat {
				counter++

				continue
			}

			counter = 0
			break
		}

		if len(matcher) == counter {
			// found a match from 'index' to 'index + counter'
			// save the address and continue searching for other matches
			newMatch := RegexMatch { Start: index, End: index + counter }
			matched = append(matched, newMatch)
		}
	}

	fmt.Println("matched string: ", matched)

	return matched
}

// If matcher == nil, match every character
func matchPartialText (text []byte, matcher []byte, starters []RegexMatch) []RegexMatch {
	matched := []RegexMatch{}

	isDotOperator := matcher == nil

	for _, starter := range starters {
		if isDotOperator {
			if starter.End < len(text) {
				newMatch := RegexMatch { Start: starter.Start, End: starter.End + 1 }
				matched = append(matched, newMatch)
			}

			continue
		}

		start := starter.End

		if start + len(matcher) > len(text) {
			continue
		}

		counter := 0
		for index, mat := range matcher {
			char := text[start + index]
			if char == mat {
				counter++
				continue
			}

			counter = 0
			break
		}

		if counter == len(matcher) {
			starter.End += counter
			newMatch := RegexMatch { Start: starter.Start, End: starter.End }

			matched = append(matched, newMatch)
		}
	}

	return matched
}

func buildParseTree(tokens []Token) (*NodeTree, []error) {
	if len(tokens) == 0 {
		return nil, nil
	}


	if len(tokens) == 1 {
		parentNode, err := parseTerminalToken(tokens)
		if err != nil {
			return parentNode, []error { err }
		}

		return parentNode, nil
	}

	// Operators ordered from lesser precedences (Binary, Unary, Group operator)
	// (in the tree, precedent op should be at the bottom)
	operators := []TokenKind{
		TOKEN_OP_OR,	// Binary Operator
		TOKEN_OP_AND,

		TOKEN_REPEAT_STAR,	// Unary Operator
		TOKEN_REPEAT_STAR_ONCE,
		TOKEN_REPEAT_PLUS,
		TOKEN_REPEAT_PLUS_ONCE,
		TOKEN_REPEAT_QUESTION,

		TOKEN_OPEN_PARENTHESE,	// Group Operator
		TOKEN_CLOSE_PARENTHESE,

		TOKEN_DOT, // Terminal Token
		TOKEN_IDENTIFIER,
	}

	var parentNode *NodeTree
	var parseErrors []error = nil
	var errs []error = nil

	foundOperatorInTokenList := false

	for _, operator := range operators {

		parentNode, errs = parseGroupOperator(operator, tokens)
		if errs != nil {
			parseErrors = append(parseErrors, errs...)
		}

		if parentNode != nil {
			foundOperatorInTokenList = true
			break
		}

		parentNode, parseErrors = parseBinaryAndUnaryOperator(operator, tokens)
		if errs != nil {
			parseErrors = append(parseErrors, errs...)
		}

		if parentNode != nil {
			foundOperatorInTokenList = true
			break
		}
	}

	if ! foundOperatorInTokenList {
		// len(tokens) > 1 && token_not_identified
		// we found no operator to apply on tokens (no binary/unary op, no group op) ???

		token := Token { Kind: TOKEN_UNKNOWN_TERMINAL, }
		for _, to := range tokens {
			to.Value = append(to.Value, []byte("--")...)
			token.Value = append(token.Value, to.Value...)
		}

		parentNode = new (NodeTree)
		parentNode.Kind = token.Kind
		parentNode.Value = token
		parentNode.Childreen = nil

		err := ParserError{ Node: parentNode, Err: ERROR_UNEXPECTED_TOKEN_FOR_PARSER }
		parseErrors = append(parseErrors, err)
	}

	return parentNode, parseErrors
}

func parseBinaryAndUnaryOperator (operator TokenKind, tokens []Token) (*NodeTree, []error) {
	var parentNode *NodeTree = nil
	var parseErrors []error = nil

	nestedGroupCount := 0

	for index := 0; index < len(tokens); index++ {
		token := tokens[index]

		if isGroupOpener(token.Kind) {
			nestedGroupCount++
		} else if isGroupCloser(token.Kind) {
			nestedGroupCount--
		}

		if operator == token.Kind && nestedGroupCount <= 0 {
			if isBinaryOperator(operator) {
				right := tokens[0:index]
				left := tokens[index+1:]

				nodeR, errRight := buildParseTree(right)
				nodeL, errLeft := buildParseTree(left)

				parentNode = new(NodeTree)
				parentNode.Kind = operator
				parentNode.Value = token
				parentNode.Childreen = []NodeTree{ *nodeR, *nodeL }

				if errRight != nil {
					parseErrors = append(parseErrors, errRight...)
				}

				if errLeft != nil {
					parseErrors = append(parseErrors, errLeft...)
				}

				break

			} else if isUnaryOperator(operator) {
				child := tokens[:index]

				// This one has problem
				nodeUnique, _ := buildParseTree(child)

				parentNode = new(NodeTree)
				parentNode.Kind = operator
				parentNode.Value = token
				parentNode.Childreen = []NodeTree { *nodeUnique }

				// Unary operator must be at the end of the token list
				if index < len(tokens) - 1 {
					val := []byte("")

					tokens := tokens[index + 1 : ]
					for _, to := range tokens {
						val = append(val, to.Value...)
					}

					token := Token { Kind: TOKEN_UNKNOWN_TERMINAL, Value: val }
					parentNode := new(NodeTree)
					parentNode.Kind = TOKEN_UNKNOWN_TERMINAL
					parentNode.Value = token

					err := ParserError{ Node: parentNode, Err: ERROR_UNARY_OPERATOR_FOUND_AT_UNEXPECTED_PLACE }
					parseErrors = append(parseErrors, err)
				}

				break

			}
		}
	}

	return parentNode, parseErrors
}

func parseGroupOperator(operator TokenKind, tokens []Token) (*NodeTree, []error) {
	var parentNode *NodeTree = nil
	var parseErrors []error = nil

	if isGroupOperator(operator) {
		indexLastElement := len(tokens) - 1
		left := tokens[0]
		right := tokens[indexLastElement]

		if isGroupOpener(left.Kind) {
			expectedCloser, err := getMatchingOperatorPair(left.Kind)

			var nodeUnique *NodeTree
			parentNode = new (NodeTree)

			if err == nil &&  expectedCloser.Kind == right.Kind {
				nodeUnique, parseErrors = buildParseTree(tokens[1:indexLastElement])
			} else {
				nodeUnique, parseErrors = buildParseTree(tokens[1:len(tokens)])
				right = *expectedCloser

				err = ParserError{ Node: parentNode, Err: ERROR_NOT_FOUND_GROUP_CLOSER_OPERATOR_PARSER }
				parseErrors = append(parseErrors, err)
			}

			val := bytes.Clone(left.Value)
			val = append(val, right.Value...)
			token := Token{ Kind: operator, Value: val }

			parentNode.Kind = operator
			parentNode.Value = token
			parentNode.Childreen = nil

			if nodeUnique != nil {
				parentNode.Childreen = []NodeTree { *nodeUnique }
			}
		}
	}

	return parentNode, parseErrors
}

func parseTerminalToken(tokens []Token) (*NodeTree, error) {
	token := tokens[0]

	parentNode := new (NodeTree)
	parentNode.Kind = token.Kind
	parentNode.Value = token
	parentNode.Childreen = nil

	if ! isTerminalNode (token.Kind) {
		err := ParserError{ Node: parentNode, Err: ERROR_EXPECTED_TERMINAL_TOKEN_FOR_PARSER }

		return parentNode, err
	}

	return parentNode, nil
}

func isTerminalNode (tk TokenKind) bool {
	if tk == TOKEN_TERMINAL || tk == TOKEN_IDENTIFIER || tk == TOKEN_DOT {
		return true
	}

	return false
}

func isBinaryOperator(op TokenKind) bool {
	if op == TOKEN_OP_OR || op == TOKEN_OP_AND {
		return true
	}

	return false
}

func isUnaryOperator(op TokenKind) bool {
	if op == TOKEN_REPEAT_PLUS || op == TOKEN_REPEAT_PLUS_ONCE {
		return true
	}

	if op == TOKEN_REPEAT_STAR || op == TOKEN_REPEAT_STAR_ONCE {
		return true
	}

	if op == TOKEN_REPEAT_QUESTION {
		return true
	}

	return false
}

func isGroupOperator(op TokenKind) bool {
	if op == TOKEN_OPEN_PARENTHESE || op == TOKEN_CLOSE_PARENTHESE {
		return true
	}

	return false
}

func isGroupOpener(op TokenKind) bool {
	if op == TOKEN_OPEN_PARENTHESE {
		return true
	}

	return false
}

func isGroupCloser(op TokenKind) bool {
	if op == TOKEN_CLOSE_PARENTHESE {
		return true
	}

	return false
}

func getMatchingOperatorPair (op TokenKind) (*Token, error) {
	if !isGroupOperator(op) {
		return nil, ERROR_MATCHING_OPERATOR_NOT_FOUND
	}

	if op == TOKEN_OPEN_PARENTHESE {
		token := Token { Kind: TOKEN_CLOSE_PARENTHESE, Value: []byte(")") }
		return &token, nil
	}

	if op == TOKEN_CLOSE_PARENTHESE {
		token := Token { Kind: TOKEN_OPEN_PARENTHESE, Value: []byte("(") }
		return &token, nil
	}

	return nil, ERROR_MATCHING_OPERATOR_NOT_FOUND
}

func tokenize(regex []byte) ([]Token, error) {
	tokens := []Token{}
	var err error = nil

	lengthRegex := len(regex)
	i := 0

	for i < lengthRegex {
		// 1. Identifier (string + number)
		if isIdentifier(regex[i]) {
			start := i 

			for i < lengthRegex && isIdentifier(regex[i]) {
				i++
			}

			token := Token { Kind: TOKEN_IDENTIFIER, Value: regex[start : i] }
			tokens = append(tokens, token)
			continue
		}

		// 2. OR, DOT, and ROUND_BRACKET
		switch regex[i] {
		case byte('|'):
			token := Token { Kind: TOKEN_OP_OR, Value: regex[i:i+1] }
			tokens = append(tokens, token)
			i++
			continue
		case byte('('):
			token := Token { Kind: TOKEN_OPEN_PARENTHESE, Value: regex[i:i+1] }
			tokens = append(tokens, token)
			i++
			continue
		case byte(')'):
			token := Token { Kind: TOKEN_CLOSE_PARENTHESE, Value: regex[i:i+1] }
			tokens = append(tokens, token)
			i++
			continue
		case byte('.'):
			token := Token { Kind: TOKEN_DOT, Value: regex[i:i+1] }
			tokens = append(tokens, token)
			i++
			continue
		}

		// 3. Frequency operation (+, *, ?)
		switch regex[i] {
		case byte('+'):
			if i + 1 < lengthRegex && regex[i + 1] == byte('?') {
				token := Token{ Kind: TOKEN_REPEAT_PLUS_ONCE, Value: regex[i:i+2]}
				tokens = append(tokens, token)
				i += 2
			} else {
				token := Token{ Kind: TOKEN_REPEAT_PLUS, Value: regex[i:i+1]}
				tokens = append(tokens, token)
				i++
			}
			continue
		case byte('*'):
			if i + 1 < lengthRegex && regex[i + 1] == byte('?') {
				token := Token{ Kind: TOKEN_REPEAT_STAR_ONCE, Value: regex[i:i+2]}
				tokens = append(tokens, token)
				i += 2
			} else {
				token := Token{ Kind: TOKEN_REPEAT_STAR, Value: regex[i:i+1]}
				tokens = append(tokens, token)
				i++
			}
			continue
		case byte('?'):
			token := Token{ Kind: TOKEN_REPEAT_QUESTION, Value: regex[i:i+1]}
			tokens = append(tokens, token)
			i++
			continue
		}

		// Catch other unexpected character type
		token := Token{ Kind: TOKEN_UNKNOWN_TERMINAL, Value: regex[i:i+1] }
		tokens = append(tokens, token)
		i++
		err = ERROR_UNEXPECTED_CHARACTER_FOR_LEXER
	}

	if len(tokens) < 2 {
		return tokens, err
	}

	// Add 'AND' operators to ease parse tree processing
	tempTokens := tokens
	tokens = []Token{}
	tokens = append(tokens, tempTokens[0])

	for i := 1; i < len(tempTokens); i++ {
		previousKind := tempTokens[i - 1].Kind
		currentKind := tempTokens[i].Kind

		if previousKind == TOKEN_IDENTIFIER && isRightOperandForAND (currentKind) {
			token := Token { Kind: TOKEN_OP_AND, Value: []byte("TOKEN_OP_AND") }
			tokens = append(tokens, token)

		} else if previousKind == TOKEN_DOT && isRightOperandForAND(currentKind) {
			token := Token { Kind: TOKEN_OP_AND, Value: []byte("TOKEN_OP_AND") }
			tokens = append(tokens, token)

		} else if previousKind == TOKEN_CLOSE_PARENTHESE && isRightOperandForAND(currentKind) {
			token := Token { Kind: TOKEN_OP_AND, Value: []byte("TOKEN_OP_AND") }
			tokens = append(tokens, token)

		} else if ( previousKind == TOKEN_REPEAT_STAR || 
		previousKind == TOKEN_REPEAT_STAR_ONCE ) && isRightOperandForAND(currentKind) {
			token := Token { Kind: TOKEN_OP_AND, Value: []byte("TOKEN_OP_AND") }
			tokens = append(tokens, token)

		} else if ( previousKind == TOKEN_REPEAT_PLUS || 
		previousKind == TOKEN_REPEAT_PLUS_ONCE ) && isRightOperandForAND(currentKind) {
			token := Token { Kind: TOKEN_OP_AND, Value: []byte("TOKEN_OP_AND") }
			tokens = append(tokens, token)

		} else if previousKind == TOKEN_REPEAT_QUESTION && isRightOperandForAND(currentKind) {
			token := Token { Kind: TOKEN_OP_AND, Value: []byte("TOKEN_OP_AND") }
			tokens = append(tokens, token)

		}

		tokens = append(tokens, tempTokens[i])
	}

	return tokens, err
}

// TODO: This function is supposed be used in the lexer
// It helps with determining the insertion for the 'AND' operator
func isRightOperandForAND(currentKind TokenKind) bool {
	if currentKind == TOKEN_OPEN_PARENTHESE {
		return true
	}

	if currentKind == TOKEN_DOT {
		return true
	}

	if currentKind == TOKEN_IDENTIFIER {
		return true
	}

	return false
}

func isIdentifier(char byte) bool {
	ok := false

	if char >= byte('a') && char <= byte('z') || char >= byte('A') && char <= byte('Z') {
		ok = true
	}

	if char >= byte('0') && char <= byte('9') {
		ok = true
	}

	switch char {
	case byte(' '):
		ok = true
	case byte(','):
		ok = true
	case byte('_'):
		ok = true
	case byte('='):
		ok = true
	// case byte('-'):
	// ok = true
	}

	return ok
}
