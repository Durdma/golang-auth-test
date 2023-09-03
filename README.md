# golang-auth-test
# Выполнение тестового задания для MEDODS

## Постановка задачи
Написать часть сервиса аутентификации.

Два REST маршрута:

- Первый маршрут выдает пару Access, Refresh токенов для пользователя с идентификатором (GUID) указанным в параметре запроса
- Второй маршрут выполняет Refresh операцию на пару Access, Refresh токенов

Стек технологий:

- Go
- JWT
- MongoDB

Требования:

- Access токен тип JWT, алгоритм SHA512, хранить в базе строго запрещено.

- Refresh токен тип произвольный, формат передачи base64, хранится в базе исключительно в виде bcrypt хеша, должен быть защищен от изменения на стороне клиента и попыток повторного использования.

- Access, Refresh токены обоюдно связаны, Refresh операцию для Access токена можно выполнить только тем Refresh токеном, который был выдан вместе с ним.

## Решение
Инструменты:
- Установленная MongoDB
- Для работы API необходим фреймвок gin: [gin-github](github.com/gin-gonic/gin)
- Для работы с JWT используется библиотека jwt-go: [jwt-go](github.com/dgrijalva/jwt-go)
- Для соединения с базой данных MongoDB необходимо использовать connector: [mongoDB driver](go.mongodb.org/mongo-driver/mongo)
- Для хэширования используется библиотека bcrypt: [bcrypt](golang.org/x/crypto/bcrypt)

Начальные настройки:
- Домен, на котором расположен API: http://localhost:8080/
- MongoDB URI: mongodb://127.0.0.1:27017
- Константы (Отнесены сюда, так как при деплое должны быть перемещены в переменные окружения)
  - timeout - предельное время выполнения операции на БД
  - mongoURI - путь подключения к БД
  - signKey - хранит специальную строку для подписи accessToken
  - AccessTokenDuration - хранит время жизни AccessToken. Время жизни должно быть достаточно коротким.
  - RefreshTokenDuration - хранит время жизни RefreshToken. Время жизни должно быть достаточно длинным.

Основные end-points:
- http://localhost:8080/auth/get-tokens?guid=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
  - http-метод запрос: POST
  - Тело запроса: пустое
  - query params:
    - guid - идентификатор пользователя. Формат: {XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}, где X может быть [0-9a-fA-F]
  - На данный end-point поступает POST запрос от пользователя, где в параметрах пути он указывает свой GUID, определенного формата. Полученный GUID валидируется и передается далее для генерации пары accessToken и refreshToken. Проверка на уникальность GUID не делается. То есть, если в БД уже есть запись о том, что существует пара токенов для передаваемого GUID, то формируется новая пара токенов, без замены старой, т. к. пользователь может использовать приложение на нескольких устройствах/браузерах. Для этого используются некоторые дополнительные параметры, которые позволяют идентифицировать приложение. Результатом обработки запроса с этого end-point является: добавление новой записи в базу данных "authDB" в коллекцию Sessions, где refreshToken хранится в хэшированном виде; Ответ от сервера с кодом 201 и телом ответа, в котором передается accessToken; в cookie записывается refreshToken в формате base64. В случае ошибки возвращается код ошибки и сообщение с текстом ошибки.
    
  - (***Updated: accessToken, при удачном выполненнии запроса, теперь также возвращается через cookie.***)
    
  - Пример запроса к end-point через curl:
    ```bash
    curl -v -X POST http://localhost:8080/auth/get-tokens?guid=123e4567-e89b-12d3-a456-9AC7CBDCEE52
    ```
  - Пример ответа при удачном выполнении:
    ```bash
    *   Trying 127.0.0.1:8080...
    * Connected to localhost (127.0.0.1) port 8080 (#0)
    > POST /auth/get-tokens?guid=123e4567-e89b-12d3-a456-9AC7CBDCEE52 HTTP/1.1
    > Host: localhost:8080
    > User-Agent: curl/8.0.1
    > Accept: */*
    >
    < HTTP/1.1 201 Created
    < Content-Type: application/json; charset=utf-8
    < Set-Cookie: refreshToken=YTk4NmI1MGI5NTQ3OTk3Mzc3ZWFhYjY2MTZhODQ4YTMwOTI1YzI5ZjNkODUxNmQ4ZmFjM2NjODAxNjk2YzViNA%3D%3D; Path=/auth;       Domain=localhost; Max-Age=1692810106; HttpOnly
    < Date: Wed, 23 Aug 2023 16:56:46 GMT
    < Content-Length: 226
    <
    {"accessToken":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTI4MDk5MjYsInN1YiI6IjEyM2U0NTY3LWU4OWItMTJkMy1hNDU2LTlBQzdDQkRDRUU1MiJ9.TRrSku1zV7J53be7MZpPH11tz9rBEheYlEnVBGBhN2NeA4ogzgk8btU7fOpCS0DGn1U8v0fn8zeW344YNvmA0Q"}* Connection #0 to host localhost left intact

  - ***Updated request***
    
    ```bash
    curl -v -X POST http://localhost:8080/auth/get-tokens?guid=12345678-4444-3333-2222-123456789222
    *   Trying 127.0.0.1:8080...
    * Connected to localhost (127.0.0.1) port 8080 (#0)
    > POST /auth/get-tokens?guid=12345678-4444-3333-2222-123456789222 HTTP/1.1
    > Host: localhost:8080
    > User-Agent: curl/8.0.1
    > Accept: */*
    >
    < HTTP/1.1 200 OK
    < Set-Cookie: refreshToken=ZmNiZDQ0MTliYWZhZjVmMjMyOTg3ZWY3NzYwYTNkOWQyZWMwN2RkYTAzODJmYzI1OWFmNDQ4MzQ2NzVhNzM2Nw%3D%3D; Path=/auth; Domain=localhost; Max-Age=300; HttpOnly
    < Set-Cookie: accessToken=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTM3NjQyODEsInN1YiI6IjEyMzQ1Njc4LTQ0NDQtMzMzMy0yMjIyLTEyMzQ1Njc4OTIyMiJ9._6fUtyW-ipB-UrJ3523YHniE9xS436pVIyHxzE3CW25ychfF9GmfFjpYJ6lSxTDn1rFiOCvK2B2h50M3H4gZsg; Path=/auth; Domain=localhost; Max-Age=120; HttpOnly
    < Date: Sun, 03 Sep 2023 18:02:41 GMT
    < Content-Length: 0
    <
    * Connection #0 to host localhost left intact
    
- http://localhost:8080/auth/refresh-tokens
  - http-метод запроса: POST
  - Тело запроса: пустое
  - На данный end-point поступает POST запрос от пользователя, где в cookies хранится refreshToken (подразумевается, что accessToken хранится в памяти приложения, поэтому обновление accessToken происходит по тому refreshToken, который был с ним выдан). Полученный refreshToken декодируется из base64 в байты. Далее выполняется поиск по БД записи, где совпадают refreshToken (для этого с помощью библиотеки bcrypt производится сравнение хэш-сумм двух токенов). Если токены совпадают, то проверяется срок действия токена. Результатом обработки запроса с этого end-point является: В случае удачных проверок токена создается новая пара accessToken и refreshToen, создается новая запись в БД (процесс аналогичен процессу добавления записи и генерации токенов на первом end-point), удаляется старая запись в БД, ответ от сервера с кодом 201 и телом ответа, в котором передается accessToken, в cookie записывается refreshToken в формате base64; В случае если токен просрочен, возвращается ошибка c кодом 403 и запись из БД с этим токеном удаляется; В случае, если запись с таким токеном не была найдена возвращается ответ с кодом ошибки 404.

  - (***Updated: Теперь на end point приходит пара cookie, в которых хранятся токены пользователя refreshToken и accessToken. Производится проверка связи между токенами, что эта пара была создана одновременно. С помощью функции CheckPairOfTokens.***)
    
  - Пример запроса к end-point через curl:
    ```bash
    curl --cookie "refreshToken=NDQyMmQ2NDE5ZDFhZjQzODZmZDg0ZGY3OWRhODA0OWY0MDlmNDRhMjZkNTExZjA2OTViMTFkYTE4Yjk3N2FmNQ==" -v -X POST http://localhost:8080/auth/refresh-tokens
    ```
  - Пример ответа при удачном выполнении:
    ```bash
    *   Trying 127.0.0.1:8080...
    * Connected to localhost (127.0.0.1) port 8080 (#0)
    > POST /auth/refresh-tokens HTTP/1.1
    > Host: localhost:8080
    > User-Agent: curl/8.0.1
    > Accept: */*
    > Cookie: refreshToken=NDQyMmQ2NDE5ZDFhZjQzODZmZDg0ZGY3OWRhODA0OWY0MDlmNDRhMjZkNTExZjA2OTViMTFkYTE4Yjk3N2FmNQ==
    >
    < HTTP/1.1 201 Created
    < Content-Type: application/json; charset=utf-8
    < Set-Cookie: refreshToken=YjQ2ZDA2ZTM3N2IwODYxY2NkZGY0YTI0YjAyZjNjMjA3NDIxMTI0YWNjYTI5MTYzNGE5NTVkNGFhNjc5YjgxNQ%3D%3D; Path=/auth; Domain=localhost; Max-Age=1692815335; HttpOnly
    < Date: Wed, 23 Aug 2023 18:23:55 GMT
    < Content-Length: 226
    <
    {"accessToken":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTI4MTUxNTUsInN1YiI6IjEyM2U0NTY3LWU4OWItMTJkMy1hNDU2LTA5ODc2NTQzMjExMSJ9.s2xNIRZ-d-Lixw33XI7I75iaFSKhQHiGGUudjqvQ3HlvykZfyvvqphWFMQuUhSuIFaVYqczPcE0-HoX5Lb7U5Q"}* Connection #0 to host localhost left intact
    ```
    
  - ***Updated пример запроса и ответа***
    
    ```bash
    curl --cookie "refreshToken=MGIwNzA3NDA5NWQxM2MwYWY5M2Q1NTRkZGRhZTg5MzEzNzBiYzhjYzhhMDJkODgyNTYzMDM5NTU3MjRhNmE1MQ%3D%3D;accessToken=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTM3NjQ1NDAsInN1YiI6IjEyMzQ1Njc4LTQ0NDQtMzMzMy0yMjIyLTEyMzQ1Njc4OTIyMiJ9.PBVniRR3uoifRGUgs19W0KlrG_NxWjC14QiTdjdJUXyW8RCj7LQdFVGarmCdUIsL3zr9wmzO7Ybg2oiZ9UrJjQ" -v -X POST http://localhost:8080/auth/refresh-tokens
    *   Trying 127.0.0.1:8080...
    * Connected to localhost (127.0.0.1) port 8080 (#0)
    > POST /auth/refresh-tokens HTTP/1.1
    > Host: localhost:8080
    > User-Agent: curl/8.0.1
    > Accept: */*
    > Cookie: refreshToken=MGIwNzA3NDA5NWQxM2MwYWY5M2Q1NTRkZGRhZTg5MzEzNzBiYzhjYzhhMDJkODgyNTYzMDM5NTU3MjRhNmE1MQ%3D%3D;accessToken=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTM3NjQ1NDAsInN1YiI6IjEyMzQ1Njc4LTQ0NDQtMzMzMy0yMjIyLTEyMzQ1Njc4OTIyMiJ9.PBVniRR3uoifRGUgs19W0KlrG_NxWjC14QiTdjdJUXyW8RCj7LQdFVGarmCdUIsL3zr9wmzO7Ybg2oiZ9UrJjQ
    >
    < HTTP/1.1 200 OK
    < Set-Cookie: refreshToken=MDgzMWViOTQzNzk5Y2FhODJhMzMyZTIxZjM5YmI4ZTg4OWIxMjk4YjUxNDFmNDI1MmMyNDUzNTE2Mzc1NGE3Nw%3D%3D; Path=/auth; Domain=localhost; Max-Age=300; HttpOnly
    < Set-Cookie: accessToken=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTM3NjQ2MDUsInN1YiI6IjEyMzQ1Njc4LTQ0NDQtMzMzMy0yMjIyLTEyMzQ1Njc4OTIyMiJ9.BYgoGgQf0fClHsVKbGPmO2ZW0mEdj8bQQq5KXfULfwWDsnUvgOLOYOah0lso4Ko_SP14H4dn5QmdPDi6SQcuJw; Path=/auth; Domain=localhost; Max-Age=120; HttpOnly
    < Date: Sun, 03 Sep 2023 18:08:05 GMT
    < Content-Length: 0
    <
    * Connection #0 to host localhost left intact
    
Структура проекта:
```
.
├── auth                # Пакет для генерации, кодирования, декодирования токенов
      └── manager.go
├── database            # Пакет для подключения и выполенния операций над БД
      └── mongo.go
├── server              # Пакет для настройки сервера
      ├── router.go     # Создание и настройка router'ов сервера
      └── server.go     # Настройка и конфигурация срвера
├── service             # Основная логика программы
      └── session.go
├── go.mod              # Основные зависимости проекта
├── go.sum              # Валидация зависимостей
└── main.go             # Запускает сущность приложения
```
