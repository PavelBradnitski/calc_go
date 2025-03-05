# Финальное задание спринта 2 YandexLMS
# Описание проекта
Данный проект представляет собой распределённый калькулятор, разработанный на языке программирования Go. Допустимые операции это + - / *.
Программа поддерживает ввод десятичных чисел с точкой в качестве разделителя. Работает по принципу преобразования выражения в постфиксную форму.
Он состоит из двух основных компонентов:

**Сервер** – управляет вычислениями, принимает выражения, раздаёт задачи агентам и сохраняет результаты.

**Агент** – получает задания от сервера, вычисляет выражения и отправляет результаты обратно.

# Установка и запуск

### Чтобы запустить программу, необходимо:
- Выполнить команду git clone git@github.com:PavelBradnitski/calc_go.git
- Перейти в 2 терминалах в созданную папку calc_go
- Выполнить команды go run ./internal/orchestrator/cmd/main.go и go run ./internal/agent/cmd/main.go в 2 терминалах
- Сервер запускается на порту 8090, если нужно это поменять используйте файл .env и переменную PORT

## Примеры работы программы:
1) Запрос с корректным выражением:
```
 curl --location --request POST 'http://localhost:8080/api/v1/calculate' \
 --header 'Content-Type: application/json' \
 --data '{
  "expression": "2+2*2"
}'
```
Код ответа 201.

2) Запрос с некорректным выражением:
```
curl --location --request POST 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "a"
}'
```
Возвращает 
```
{
    "error": "Expression is not valid"
} 
```
Код ответа 422.

3) Получение списка всех выражений: 
```
 curl --location --request GET 'http://localhost:8090/api/v1/expressions' \
 --header 'Content-Type: application/json'
```

Тело ответа:
```
{"expressions":{"id":0,"status":"done","result":6}}
```
Код ответа 200.

4) Получение выражения по id: 
```
 curl --location --request GET 'http://localhost:8090/api/v1/expressions/?id=0' \
 --header 'Content-Type: application/json'
```

Тело ответа:
```
{"expressions":{"id":0,"status":"done","result":6}}
```
Код ответа 200.

5) Пример запроса, где ID не найден
```
 curl --location --request GET 'http://localhost:8090/api/v1/expressions/?id=1' \ 
 --header 'Content-Type: application/json'
```

Тело ответа:
```
{"Error": "not found by id"}
```
Код ответа 404.