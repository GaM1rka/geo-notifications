# Geo Notifications Service

Сервис для управления гео‑инцидентами и проверки, какие инциденты актуальны для пользователя по его геолокации. Реализован на Go, использует PostgreSQL и Redis, поддерживает интеграцию с внешними вебхуками через ngrok.

## Стек

- Go
- PostgreSQL
- Redis
- Docker и docker-compose
- Ngrok (для проброса вебхуков наружу)

## Структура проекта (общая идея)

- `cmd/app` — основной HTTP‑сервер
- `internal/handler` — HTTP‑хендлеры
- `internal/service` — бизнес‑логика (`IncidentService`)
- `internal/repository` — работа с PostgreSQL и Redis
- `internal/model` — модели данных
- `cmd/webhook-mock` — моковый вебхук‑сервер

Названия директорий могут отличаться, но логика разделения слоёв примерно такая.

## Настройка окружения

Создайте файл `.env` в корне проекта и опишите основные переменные:

```env
# База данных
DATABASE_URL=postgres://postgres:postgres@postgres:5432/geo?sslmode=disable

# Redis
REDIS_ADDR=redis:6379

# Дополнительные параметры при необходимости
# WEBHOOK_URL=скопировать и вставить из ngrok
```
## Запуск через Docker
``` bash
docker-compose up --build -d
```
будут запущены:
- приложение (geo-notifications-app);
- PostgreSQL (postgres);
- Redis (redis);

## Проверка health-check
``` bash
curl http://localhost:8080/api/v1/system/health
```
ожидает ответ:
```json
{
  "status": "ok",
  "db": "ok",
  "redis": "ok"
}
```

## Основные HTTP‑эндпоинты
Пример тела запроса:
POST /incidents
```json
{
  "title": "Road accident",
  "description": "Accident near the bridge",
  "latitude": 55.75,
  "longitude": 37.61,
  "radius_m": 500
}
```
Ответ при успехе: 201 Created и JSON c созданным инцидентом.

GET /incidents — список инцидентов с пагинацией.
Поддерживаемые query‑параметры:
page — номер страницы (по умолчанию 1);
page_size — размер страницы (по умолчанию 20).
Пример:
``` bash
curl "http://localhost:8080/api/v1/incidents?page=1&page_size=20"
```
Ответ:
```json
{
  "items": [
    {
      "id": 1,
      "title": "Road accident",
      "description": "Accident near the bridge",
      "latitude": 55.75,
      "longitude": 37.61,
      "radius_m": 500,
      "active": true
    }
  ],
  "page": 1,
  "page_size": 20
}
```

GET /incidents/{id} — получить инцидент по идентификатору.

PUT /incidents/{id} — обновить инцидент (тело аналогично созданию; ID берётся из пути).

DELETE /incidents/{id} — деактивировать (логически удалить) инцидент.

GET /incidents/stats — возвращает количество уникальных пользователей за последнее окно в N минут.

Значение N задаётся при инициализации хендлера (поле statsWindowMinutes).

Возвращаемый JSON:
```json
{
  "user_count": 42
}
```
## Моковый вебхук‑сервер и Ngrok
# Запускаем mock сервер:
``` bash
go run ./cmd/webhook-mock/main.go
```
# Проброс порта через Ngrok
Чтобы внешний сервис мог отправлять вебхуки на ваш локальный мок‑сервер:
Установите и залогиньтесь в ngrok.
В отдельном терминале выполните:
``` bash
ngrok http 9090
```
Ngrok выдаст публичный URL вида:
``` text
https://random-subdomain.ngrok.io
```
Этот URL используйте как внешний webhook‑URL:
либо пропишите его в .env, например:
``` text
WEBHOOK_URL=https://random-subdomain.ngrok.io/webhook
```
## Тестирование
# Юнит‑тесты
Юнит‑тесты для HTTP‑хендлеров живут в internal/handler и используют:
стандартный пакет net/http/httptest для имитации запросов/ответов;
фейковую реализацию IncidentService, не трогающую реальную базу и Redis.
Запуск:

``` bash
go test ./internal/handler -count=1 -v
```
# Интеграционные тесты
Интеграционные тесты используют реальные DATABASE_URL и REDIS_ADDR. Перед запуском:
Поднимите PostgreSQL и Redis (через Docker или локально).
Экспортируйте нужные переменные окружения:

``` bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/geo?sslmode=disable"
export REDIS_ADDR="localhost:6379"
```
Запустите тесты:

``` bash
go test ./... -count=1 -v
```
