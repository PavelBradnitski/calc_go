package agent

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PavelBradnitski/calc_go/internal/orchestrator"
)

func AgentRun() {
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
				resp, err := http.Get("http://localhost:8090/internal/task")
				if err != nil {
					log.Println("Ошибка получения задачи:", err)
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

				respPost, err := http.Post("http://localhost:8090/internal/result", "application/json", strings.NewReader(string(jsonData)))
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
		time.Sleep(time.Duration(GetExecTime("TIME_ADDITION_MS")) * time.Microsecond)
		return task.Arg1 + task.Arg2
	case "-":
		time.Sleep(time.Duration(GetExecTime("TIME_SUBTRACTION_MS")) * time.Microsecond)
		return task.Arg1 - task.Arg2
	case "*":
		time.Sleep(time.Duration(GetExecTime("TIME_MULTIPLICATIONS_MS")) * time.Microsecond)
		return task.Arg1 * task.Arg2
	case "/":
		time.Sleep(time.Duration(GetExecTime("TIME_DIVISIONS_MS")) * time.Microsecond)
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}
func GetExecTime(env string) int {
	val, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		return 2
	}
	return val
}
