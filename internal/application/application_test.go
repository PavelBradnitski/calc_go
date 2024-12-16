package application

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalcHandler(t *testing.T) {
	tests := []struct {
		name             string
		expression       string
		wantedStatusCode int
		wantError        bool
		wantedResult     float64
	}{
		{
			name:             "simple",
			expression:       "1+1+1",
			wantedStatusCode: http.StatusOK,
			wantedResult:     3,
			wantError:        false,
		},
		{
			name:             "priority",
			expression:       "(2+2)*2",
			wantedStatusCode: http.StatusOK,
			wantedResult:     8,
			wantError:        false,
		},
		{
			name:             "priority",
			expression:       "2+2*2",
			wantedStatusCode: http.StatusOK,
			wantedResult:     6,
			wantError:        false,
		},
		{
			name:             "/",
			expression:       "-1/2",
			wantedStatusCode: http.StatusOK,
			wantedResult:     -0.5,
			wantError:        false,
		},
		{
			name:             "division by zero",
			expression:       "2/0",
			wantedStatusCode: http.StatusBadRequest,
			wantedResult:     0,
			wantError:        true,
		},
		{
			name:             "invalid charachter",
			expression:       "a",
			wantedStatusCode: http.StatusBadRequest,
			wantedResult:     0,
			wantError:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonExpression := fmt.Sprintf("{\"expression\": \"%s\"}", tt.expression)
			request, _ := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(jsonExpression)))
			response := httptest.NewRecorder()
			CalcHandler(response, request)
			got := response.Body.String()
			expected := fmt.Sprintf("result: %f", tt.wantedResult)
			if !tt.wantError {
				if got != expected {
					t.Errorf("got %s,want %s", got, expected)
				}
			} else {
				if tt.wantedStatusCode != response.Code {
					t.Errorf("wrong status code. got %s,want %s", got, expected)
				}
			}

		})
	}
}
