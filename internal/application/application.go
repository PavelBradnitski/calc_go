package application

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

// Функция запуска приложения
// тут будем читать введенную строку и после нажатия ENTER писать результат работы программы на экране
// если пользователь ввел exit - то останаваливаем приложение
func (a *Application) Run() error {
	for {
		// читаем выражение для вычисления из командной строки
		log.Println("input expression")
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("failed to read expression from console")
		}
		// убираем пробелы, чтобы оставить только вычислемое выражение
		text = strings.TrimSpace(text)
		// выходим, если ввели команду "exit"
		if text == "exit" {
			log.Println("aplication was successfully closed")
			return nil
		}
		// Проверяем правильность введенного выражения и приводим выражение
		// из string в []string, отделяя числа, скобки, знаки друг от друга
		if expressionInSlice, err := calculation.ParseExpression(text); err != nil {
			log.Println(text, " parsing expression failed wit error: ", err)
		} else {
			//вычисляем выражение
			result, err := calculation.Calculator(expressionInSlice)
			if err != nil {
				log.Println(text, " calculation failed wit error: ", err)
			} else {
				log.Println(text, "=", result)
			}
		}
	}
}

type Request struct {
	Expression string `json:"expression"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrMethod)
		return
	}
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
	result, err := calculation.Calculator(expressionInSlice)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrInvalidExpression)
		return
	} else {
		fmt.Fprintf(w, "{\n\tresult: \"%f\"\n}", result)
	}
}

func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
