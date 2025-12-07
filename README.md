### Что сделано
- Реализована система аутентификации и авторизации по вашему ТЗ на Go:
    - Веб‑фреймворк: Gin.
    - БД: SQLite (через GORM), авто‑миграция модели `User`.
    - Пароли: `bcrypt` (hash/verify).
    - Токены: JWT (HS256) с полями `user_id`, `role`, `exp`.
    - Middleware: проверка JWT и ролевая авторизация (например, только `admin`).
    - Эндпоинты: `POST /register`, `POST /login`, `GET /profile`, `PUT /profile`, `DELETE /user/:id` (admin).
    - Конфигурация через переменные окружения: `PORT`, `JWT_SECRET`, `DB_PATH`.

### Структура проекта
```
/cmd/server/main.go               — запуск HTTP сервера
/internal/database/sqlite.go      — инициализация GORM + SQLite
/models/user.go                   — модель пользователя
/handlers/handlers.go             — HTTP обработчики
/middleware/auth.go               — JWT‑аутентификация
/middleware/role.go               — проверка ролей
/pkg/config/config.go             — загрузка конфигурации из ENV
/pkg/security/password.go         — bcrypt хеширование/проверка
/main.go                          — подсказка по запуску
```

### Конфигурация (ENV)
- `PORT` — порт сервера (по умолчанию `8080`).
- `JWT_SECRET` — секрет для подписи JWT (обязательно переопределить в проде!).
- `DB_PATH` — путь к SQLite файлу (по умолчанию `app.db`).

Можно использовать файл `.env` в корне проекта (загружается автоматически):

```
PORT=8080
JWT_SECRET=super-secret-change-me
DB_PATH=app.db
```

### Как запустить
1) Установите переменные окружения (минимум `JWT_SECRET`):
```
export JWT_SECRET="super-secret-change-me"
export PORT=8080
export DB_PATH=app.db
```
2) Запустите сервер:
```
go run ./cmd/server
```

Примечание: при первом запуске выполните `go mod tidy`, чтобы подтянуть зависимости (`github.com/joho/godotenv`).

### Примеры запросов
- Регистрация:
```
curl -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"s3cret","role":"admin"}'
```
- Логин (получить JWT):
```
curl -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"s3cret"}'
```
Ответ: `{ "token": "<JWT>" }`

- Профиль (GET):
```
TOKEN=<JWT>
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/profile
```
- Обновление профиля (PUT):
```
curl -X PUT http://localhost:8080/profile \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"password":"newpass"}'
```
- Удаление пользователя (admin):
```
curl -X DELETE http://localhost:8080/user/2 \
  -H "Authorization: Bearer $TOKEN"
```

### Безопасность и рекомендации
- Используйте HTTPS в продакшене (TLS‑сертификат, обратный прокси nginx/caddy/traefik).
- Обязательно задайте сильный `JWT_SECRET` через ENV и храните секреты безопасно.
- По желанию можно добавить refresh‑token поток и ротацию токенов (готов расширить по запросу).

### Совместимость
- Модуль `go 1.22.5` и зависимости добавлены в `go.mod`. Сборка: `go run ./cmd/server`.