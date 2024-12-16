package calculation

import "errors"

var (
	ErrInvalidExpression = errors.New("Expression is not valid")
	ErrInternalServer    = errors.New("Internal server error")
	ErrDivisionByZero    = errors.New("division by zero")
	ErrInvidCharachter   = errors.New("invalid charachter")
	ErrBracket           = errors.New("bracket error")
	ErrArithmeticSign    = errors.New("incorrect use arithmetic sign")
	ErrPostfixExpression = errors.New("invalid postfix expression")
)
