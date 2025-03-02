package calculation

import (
	"testing"
)

func TestCalculator(t *testing.T) {
	testCasesSuccess := []struct {
		name           string
		expression     string
		expectedResult float64
		wantError      bool
	}{
		{
			name:           "simple",
			expression:     "1+1+1",
			expectedResult: 3,
			wantError:      false,
		},
		{
			name:           "priority",
			expression:     "(2+2)*2",
			expectedResult: 8,
			wantError:      false,
		},
		{
			name:           "priority",
			expression:     "2+2*2",
			expectedResult: 6,
			wantError:      false,
		},
		{
			name:           "/",
			expression:     "1/2",
			expectedResult: 0.5,
			wantError:      false,
		},
		{
			name:           "simple invalid",
			expression:     "2/0",
			expectedResult: 0,
			wantError:      true,
		},
	}
	for _, testCase := range testCasesSuccess {
		t.Run(testCase.name, func(t *testing.T) {
			input, err := ParseExpression(testCase.expression)
			if err != nil {
				if !testCase.wantError {
					t.Fatalf("expected an error")
				}
			}
			val, err := Calculator(input)
			if err != nil && testCase.wantError {
				t.Fatalf("not expected error")
			}
			result, err := CalculatePrefix(*val)
			if testCase.wantError {
				if err == nil {
					t.Fatalf("expected an error")
				}
			} else {
				if err != nil {
					t.Fatalf("successful case %s returns error", testCase.expression)
				}
				if result != testCase.expectedResult {
					t.Fatalf("%f should be equal %f", result, testCase.expectedResult)
				}
			}
		})
	}
}
