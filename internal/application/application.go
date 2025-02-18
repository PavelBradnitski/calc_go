package application

import (
	"fmt"
	"log"
	"os"

	"github.com/PavelBradnitski/calc_go/http/server/handler"
	"github.com/PavelBradnitski/calc_go/internal/repositories"
	"github.com/PavelBradnitski/calc_go/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate"
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

// func CalcHandler(w http.ResponseWriter, r *http.Request) {
// 	request := new(Request)
// 	defer r.Body.Close()
// 	err := json.NewDecoder(r.Body).Decode(&request)
// 	if r.Method != http.MethodPost {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrMethod)
// 		return
// 	}
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrInternalServer)
// 		return
// 	}
// 	expressionInSlice, err := calculation.ParseExpression(request.Expression)
// 	if err != nil {
// 		w.WriteHeader(http.StatusUnprocessableEntity)
// 		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrInvalidExpression)
// 		return
// 	}
// 	id, err := calculation.Calculator(expressionInSlice)
// 	if err != nil {
// 		w.WriteHeader(http.StatusUnprocessableEntity)
// 		fmt.Fprintf(w, "{\n\terror: \"%s\"\n}", calculation.ErrInvalidExpression)
// 		return
// 	} else {
// 		w.WriteHeader(http.StatusCreated)
// 		fmt.Fprintf(w, "{\n\tid: \"%d\"\n}", id)
// 	}
// }

func (a *Application) RunServer() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	// Подключение к БД
	db, err := repositories.ConnectToDB(dbUser, dbPassword, dbHost, dbPort, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer db.Close()

	// Запуск миграции
	connectionString := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	m, err := migrate.New(
		"file:///Rates/db/migrations",
		connectionString,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations applied successfully!")
	rateRepo := repositories.NewRateRepository(db)

	// Создание HTTP сервера
	rateService := services.NewRateService(rateRepo)
	rateHandler := handler.NewRateHandler(rateService)
	router := gin.Default()
	rateHandler.RegisterRoutes(router)
	go router.Run(":8080")
}
