# calc_go
Финальное задание спринта 1 YandexLMS

    Данная программа позволяет вычислить введенное корректное арифметическое выражение. Допустимые операции это + - / *.
Программа поддерживает ввод десятичных чисел с точкой в качестве разделителя. Работает по принципу преобразования выражения в постфиксную форму.
Есть 2 версии работы приложения: 
 - Консоль. на вход арифметическое выражение в виде строки. 
    Чтобы запустить необходимо раскомментировать app.Run() в main.go и закомментировать app.RunServer().
 - Сервер. на вход арифметическое выражение в виде JSON. Пример:
{
    "expression": "выражение, которое ввёл пользователь"
}

Чтобы запустить программу, необходимо:
1) Выполнить команду git clone git@github.com:PavelBradnitski/calc_go.git
2) Перейти в созданную папку calc_go
3) Выполнить команду go run ./cmd/main.go

Примеры работы программы:
1) curl --location --request POST 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
Возвращает
{
    "result": "6.000000"
}
Код 200.
2) curl --location --request POST 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "a"
}'
Возвращает 
{
    "error": "Expression is not valid"
} 
Код 422.
3) curl --location --request POST 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json'
Возвращает 
{
    "error": "Internal server error"
}
Код 500.
