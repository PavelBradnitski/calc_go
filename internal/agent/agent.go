package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PavelBradnitski/calc_go/internal/orchestrator"
	"github.com/joho/godotenv"
)

func AgentRun() {
	err := godotenv.Load()

	if err != nil {
		log.Printf("Failed to open .env")
	}
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Printf("Port is not a number")
		port = 8090
	}
	path := fmt.Sprintf("http://localhost:%v/internal/task", port)
	computingPower, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if computingPower == 0 {
		computingPower = 2
	}

	for i := 0; i < computingPower; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Panic recovered:", r)
				}
			}()

			for {

				resp, err := http.Get(path)
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				}

				if resp.StatusCode != http.StatusOK {
					time.Sleep(1 * time.Second) // Не спамим сервер
					resp.Body.Close()
					continue
				}

				log.Println("Задача получена, обрабатываем...")

				var response struct {
					Task orchestrator.Task `json:"task"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					resp.Body.Close()
					log.Println("Ошибка декодирования JSON:", err)
					continue
				}

				result := computeOperation(response.Task)
				jsonData, _ := json.Marshal(orchestrator.Result{
					ExpressionID: response.Task.ExpressionID,
					TaskID:       response.Task.ID,
					Result:       result,
				})

				log.Println("Отправляем результат:", string(jsonData))

				respPost, err := http.Post("http://localhost:8090/internal/task", "application/json", strings.NewReader(string(jsonData)))
				if err != nil {
					log.Println("Ошибка отправки результата:", err)
					continue
				}
				respPost.Body.Close()

				log.Println("Результат отправлен, статус ответа:", respPost.StatusCode)
			}
		}()
	}

	select {} // Блокируем выполнение, пока не завершится программа
}

func computeOperation(task orchestrator.Task) float64 {
	switch task.Operation {
	case "+":
		time.Sleep(time.Duration(orchestrator.GetExecTime("TIME_ADDITION_MS")) * time.Millisecond)
		return task.Arg1 + task.Arg2
	case "-":
		time.Sleep(time.Duration(orchestrator.GetExecTime("TIME_SUBTRACTION_MS")) * time.Millisecond)
		return task.Arg1 - task.Arg2
	case "*":
		time.Sleep(time.Duration(orchestrator.GetExecTime("TIME_MULTIPLICATIONS_MS")) * time.Millisecond)
		return task.Arg1 * task.Arg2
	case "/":
		time.Sleep(time.Duration(orchestrator.GetExecTime("TIME_DIVISIONS_MS")) * time.Millisecond)
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}
