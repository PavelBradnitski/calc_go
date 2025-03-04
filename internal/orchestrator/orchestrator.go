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
	"github.com/joho/godotenv"
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
	ExprIndex   int
	ResultChan  chan Result
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		Expressions: make(map[int]Expression),
		Results:     make(map[int]float64),
		ExprIndex:   0,
		ResultChan:  make(chan Result, 100),
	}
}

func StartOrchestrator() {
	orchestrator := NewOrchestrator()
	go orchestrator.ProcessResults()

	http.HandleFunc("/api/v1/calculate", orchestrator.AddExpression)
	http.HandleFunc("/api/v1/expressions", orchestrator.GetExpressions)
	http.HandleFunc("/internal/task", orchestrator.HandleTask)

	fmt.Println("Server is running on :8090")
	fmt.Println("Registered routes:")
	fmt.Println("- POST /api/v1/calculate")
	fmt.Println("- GET/POST /internal/task")

	err := godotenv.Load()
	if err != nil {
		log.Printf("Failed to open .env")
	}

	log.Fatal(http.ListenAndServe(":8090", nil))
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
	// Разбираем выражение в задачи
	go func(exprID int, rpn *[]string, orchestrator *Orchestrator) {
		orchestrator.ParseExpressionToTasks(exprID, *rpn)
	}(id, postfix, o)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (o *Orchestrator) ParseExpressionToTasks(exprID int, postfixExpr []string) {
	var stack []float64
	taskID := 0
	fmt.Println("🔄 Starting ParseExpressionToTasks for ID:", exprID)

	for _, token := range postfixExpr {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
			continue
		}

		if len(stack) < 2 {
			log.Println("Ошибка: недостаточно операндов для операции", token)
			return
		}

		arg2 := stack[len(stack)-1]
		arg1 := stack[len(stack)-2]
		stack = stack[:len(stack)-2]

		execTime := getExecTimeForOp(token)

		task := Task{
			ExpressionID: exprID,
			ID:           exprID*100 + taskID,
			Arg1:         arg1,
			Arg2:         arg2,
			Operation:    token,
			ExecTime:     execTime,
		}
		fmt.Printf("📌 Created task: %+v\n", task)
		o.RWMutex.Lock()
		o.Tasks = append(o.Tasks, task)
		o.RWMutex.Unlock()
		result := o.WaitForTaskResult(task.ExpressionID, task.ID)

		stack = append(stack, result)

		taskID++
	}
	// Проверяем, что в конце в стеке осталось одно значение (результат)
	if len(stack) != 1 {
		log.Println("Ошибка: некорректное постфиксное выражение")
		return
	}
	fmt.Printf("🏁 Result %v\n", stack[0])
	o.RWMutex.Lock()
	o.Expressions[exprID] = Expression{ID: exprID, Status: "done", Result: stack[0]}
	o.RWMutex.Unlock()
}
func (o *Orchestrator) WaitForTaskResult(expID, taskID int) float64 {
	for {
		o.RWMutex.RLock()
		result, exists := o.Expressions[expID].SubResults[taskID]
		o.RWMutex.RUnlock()

		if exists {
			return result
		}

		time.Sleep(1 * time.Second) // Ждем, чтобы не грузить процессор
	}
}

func getExecTimeForOp(op string) int {
	switch op {
	case "+":
		return GetExecTime("TIME_ADDITION_MS")
	case "-":
		return GetExecTime("TIME_SUBTRACTION_MS")
	case "*":
		return GetExecTime("TIME_MULTIPLICATIONS_MS")
	case "/":
		return GetExecTime("TIME_DIVISIONS_MS")
	default:
		return 5000
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

func (o *Orchestrator) HandleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		o.GetTask(w, r)
	} else if r.Method == http.MethodPost {
		o.ReceiveResult(w, r)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GetTask отдает агенту следующую задачу, удаляя её из списка
func (o *Orchestrator) GetTask(w http.ResponseWriter, r *http.Request) {
	o.RWMutex.Lock()
	defer o.RWMutex.Unlock()

	if len(o.Tasks) == 0 {
		http.Error(w, "No tasks available", http.StatusNotFound)
		return
	}

	task := o.Tasks[0]
	o.Tasks = o.Tasks[1:]

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
		o.RWMutex.RUnlock()
	}
}

func GetExecTime(env string) int {
	val, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		return 5000
	}
	return val
}
