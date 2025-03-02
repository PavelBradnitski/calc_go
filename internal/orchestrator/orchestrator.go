package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/PavelBradnitski/calc_go/pkg/calculation"
)

type Expression struct {
	ID         int             `json:"id"`
	Status     string          `json:"status"`
	Result     float64         `json:"result,omitempty"`
	SubResults map[int]float64 `json:"-"` // Сюда сохраняем промежуточные результаты
}

type Task struct {
	ExpressionID int     `json:"expId"`
	ID           int     `json:"id"`
	Arg1         float64 `json:"arg1"`
	Arg2         float64 `json:"arg2"`
	Operation    string  `json:"operation"`
	ExecTime     int     `json:"operation_time"`
}

type Result struct {
	ExpressionID int     `json:"expId"`
	TaskID       int     `json:"taskId"`
	Result       float64 `json:"result"`
}

type Orchestrator struct {
	Expressions map[int]Expression
	Tasks       []Task
	Results     map[int]float64
	RWMutex     sync.RWMutex
	TaskIndex   int
	ExprIndex   int
	ResultChan  chan Result
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		Expressions: make(map[int]Expression),
		Results:     make(map[int]float64),
		TaskIndex:   0,
		ExprIndex:   0,
		ResultChan:  make(chan Result, 100),
	}
}

func (o *Orchestrator) AddExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("✅ Received POST /api/v1/calculate")

	var data struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		fmt.Println("❌ JSON decode error:", err)
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}
	fmt.Println("📥 Expression received:", data.Expression)
	// Генерируем ID для выражения
	o.RWMutex.Lock()
	id := o.ExprIndex
	o.ExprIndex++
	o.Expressions[id] = Expression{
		ID:         id,
		Status:     "pending",
		SubResults: make(map[int]float64),
	}
	o.RWMutex.Unlock()
	expressionInSlice, err := calculation.ParseExpression(data.Expression)
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
	go func(exprID int, rpn *[]string, orchestrator *Orchestrator) {
		orchestrator.ParseExpressionToTasks(exprID, *rpn)
	}(id, postfix, o)
	// Разбираем выражение в задачи

	// Отправляем ID клиенту
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (o *Orchestrator) ParseExpressionToTasks(exprID int, postfixExpr []string) {
	//var tasks []Task
	var stack []float64
	taskID := 0
	fmt.Println("🔄 Starting ParseExpressionToTasks for ID:", exprID)

	for _, token := range postfixExpr {
		// Если это число → кладем в стек
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
			continue
		}

		// Это оператор, значит, достаем два операнда
		if len(stack) < 2 {
			log.Println("Ошибка: недостаточно операндов для операции", token)
			return
		}

		// Достаем операнды из стека (arg2 - последний, arg1 - предпоследний)
		arg2 := stack[len(stack)-1]
		arg1 := stack[len(stack)-2]
		stack = stack[:len(stack)-2] // Удаляем их из стека

		// Определяем время выполнения
		execTime := getExecTimeForOp(token)

		// Создаем задачу
		task := Task{
			ExpressionID: exprID,
			ID:           exprID*100 + taskID, // Генерируем уникальный ID для задачи
			Arg1:         arg1,
			Arg2:         arg2,
			Operation:    token,
			ExecTime:     execTime,
		}
		fmt.Println("📌 Created task:", task)
		o.RWMutex.Lock()
		o.Tasks = append(o.Tasks, task)
		o.RWMutex.Unlock()
		fmt.Printf("Tasks %v\n", o.Tasks)
		// Ждем выполнения задачи и получаем результат
		result := o.WaitForTaskResult(task.ExpressionID, task.ID)
		fmt.Printf("Result %v\n", result)
		// Результат этой операции кладем обратно в стек
		stack = append(stack, result)

		taskID++
	}
	fmt.Printf("Finished\n")

	// Проверяем, что в конце в стеке осталось одно значение (результат)
	if len(stack) != 1 {
		log.Println("Ошибка: некорректное постфиксное выражение")
		return
	}
	o.RWMutex.Lock()
	o.Expressions[exprID] = Expression{ID: exprID, Status: "done", Result: stack[0]}
	o.RWMutex.Unlock()
}
func (o *Orchestrator) WaitForTaskResult(expID, taskID int) float64 {
	for {
		o.RWMutex.RLock() // 🔒 Разрешаем множественное чтение
		result, exists := o.Expressions[expID].SubResults[taskID]
		o.RWMutex.RUnlock() // 🔓 Освобождаем чтение

		if exists {
			return result
		}

		time.Sleep(1 * time.Second) // Ждем, чтобы не грузить процессор
	}
}

// Функция для определения времени выполнения операции
func getExecTimeForOp(op string) int {
	switch op {
	case "+":
		return getExecTime("TIME_ADDITION_MS")
	case "-":
		return getExecTime("TIME_SUBTRACTION_MS")
	case "*":
		return getExecTime("TIME_MULTIPLICATIONS_MS")
	case "/":
		return getExecTime("TIME_DIVISIONS_MS")
	default:
		return 100 // Значение по умолчанию
	}
}

func (o *Orchestrator) GetExpressions(w http.ResponseWriter, r *http.Request) {
	var expressions []Expression
	o.RWMutex.RLock()
	for _, expr := range o.Expressions {
		expressions = append(expressions, expr)
	}
	o.RWMutex.RUnlock()
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions})
}

//	func (o *Orchestrator) HandleTask(w http.ResponseWriter, r *http.Request) {
//		if r.Method == http.MethodGet {
//			o.GetTask(w, r)
//		} else if r.Method == http.MethodPost {
//			o.ReceiveResult(w, r)
//		} else {
//			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
//		}
//	}
//
// GetTask отдает агенту следующую задачу, удаляя её из списка
func (o *Orchestrator) GetTask(w http.ResponseWriter, r *http.Request) {
	o.RWMutex.Lock() // 🔒 Блокируем на запись, чтобы избежать гонки данных
	defer o.RWMutex.Unlock()

	if len(o.Tasks) == 0 {
		http.Error(w, "No tasks available", http.StatusNotFound)
		return
	}

	task := o.Tasks[0]
	o.Tasks = o.Tasks[1:] // Удаляем задачу из очереди

	json.NewEncoder(w).Encode(map[string]Task{"task": task})
}

func (o *Orchestrator) ReceiveResult(w http.ResponseWriter, r *http.Request) {
	var result Result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	o.RWMutex.Lock()
	defer o.RWMutex.Unlock()

	o.ResultChan <- result

	w.WriteHeader(http.StatusOK)
}

func (o *Orchestrator) ProcessResults() {
	for result := range o.ResultChan {
		o.RWMutex.RLock()
		o.Results[result.ExpressionID] = result.Result
		o.Expressions[result.ExpressionID].SubResults[result.TaskID] = result.Result
		o.Expressions[result.ExpressionID] = Expression{ID: result.ExpressionID, Status: "completed", Result: result.Result, SubResults: o.Expressions[result.ExpressionID].SubResults}
		//o.Expressions[result.ExpressionID] = Expression{ID: result.ExpressionID, Status: "completed", Result: result.Result, SubResults: map[int]float64{result.TaskID: result.Result}}
		o.RWMutex.RUnlock()
	}
}

func StartOrchestrator() {
	orchestrator := NewOrchestrator()
	go orchestrator.ProcessResults()

	http.HandleFunc("/api/v1/calculate", orchestrator.AddExpression)
	http.HandleFunc("/api/v1/expressions", orchestrator.GetExpressions)
	http.HandleFunc("/internal/task", orchestrator.GetTask)
	http.HandleFunc("/internal/result", orchestrator.ReceiveResult)

	fmt.Println("Server is running on :8090")
	fmt.Println("Registered routes:")
	fmt.Println("- POST /api/v1/calculate")
	fmt.Println("- GET/POST /internal/task")

	// // Логируем все входящие запросы
	// loggedMux := http.NewServeMux()
	// loggedMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("Received request:", r.Method, r.URL.Path)
	// 	http.DefaultServeMux.ServeHTTP(w, r)
	// })

	log.Fatal(http.ListenAndServe(":8090", nil))
}

func getExecTime(env string) int {
	val, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		return 100
	}
	return val
}
