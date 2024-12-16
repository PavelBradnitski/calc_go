# calc_go
Финальное задание спринта 1 YandexLMS

    Данная программа позволяет вычислить введенное корректное арифметическое выражение. Допустимые операции это + - / *.
Программа поддерживает ввод десятичных чисел с точкой в качестве разделителя. Работает по принципу преобразования выражения в постфиксную форму.
Есть 2 версии работы приложения: 
 - Консоль. На вход арифметическое выражение в виде строки. 
    Чтобы запустить необходимо раскомментировать app.Run() в main.go и закомментировать app.RunServer().
 - Сервер. На вход арифметическое выражение в виде JSON. Пример:
{
    "expression": "выражение, которое ввёл пользователь"
}

Чтобы запустить программу, необходимо:
1) Запустить команду git clone git@github.com:PavelBradnitski/calc_go.git
2) Перейти в созданную папку calc_go
3) Ввести команду go run ./cmd/main.go

Примеры работы программы:
1) curl --request POST --header "Content-Type: application/json" --data "{ \"expression\": \"2+2*2\" }" http://localhost:8080/api/v1/calculate
Возвращает
{
    "result": "результат выражения"
}
Код 200.
2) curl --request POST --header "Content-Type: application/json" --data "{ \"expression\": \"a\" }" http://localhost:8080/api/v1/calculate
Возвращает 
{
    "error": "Expression is not valid"
} 
Код 422.
3) curl --request POST --header "Content-Type: application/json" --data "" http://localhost:8080/api/v1/calculate
Возвращает 
{
    "error": "Internal server error"
}
Код 500.
