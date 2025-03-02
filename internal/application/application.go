package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/PavelBradnitski/calc_go/internal/agent"
	"github.com/PavelBradnitski/calc_go/internal/orchestrator"
	"github.com/PavelBradnitski/calc_go/pkg/calculation"
)

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

type Request struct {
	Expression string `json:"expression"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrInternalServer)
		return
	}
	expressionInSlice, err := calculation.ParseExpression(request.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrInvalidExpression)
		return
	}
	postfix, err := calculation.Calculator(expressionInSlice)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrInvalidExpression)
		return
	}
	result, err := calculation.CalculatePrefix(*postfix)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrInvalidExpression)
		return
	} else {
		fmt.Fprintf(w, "{\n\tresult: \"%f\"\n}", result)
	}
}

func (a *Application) RunServer() {
	go agent.AgentRun()
	orchestrator.StartOrchestrator()
}
