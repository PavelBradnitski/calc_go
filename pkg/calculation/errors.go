package calculation

import "errors"

var (
	ErrInvalidExpression = errors.New("invalid expression")
	ErrDivisionByZero    = errors.New("division by zero")
	ErrInvidCharachter   = errors.New("invalid charachter")
	ErrBracket           = errors.New("bracket error")
	ErrArithmeticSign    = errors.New("incorrect use arithmetic sign")
	ErrPostfixExpression = errors.New("invalid postfix expression")
)
