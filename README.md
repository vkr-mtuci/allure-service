# Allure Service

Этот сервис предназначен для интеграции с Allure API. Он предоставляет REST API для получения информации о запусках тестов, генерации отчетов и их скачивания.

## 📌 Возможности
- Получение информации о ближайшем запуске тестов после указанной даты.
- Генерация PDF-отчета по результатам тестирования.
- Скачивание PDF-отчета напрямую с бэкенда.
- Логирование запросов и ошибок.
- Гибкая конфигурация через переменные окружения.

## 🚀 Технологии
- **Язык**: Go
- **Фреймворк**: Fiber (gofiber.io)
- **HTTP-клиент**: resty (go-resty/resty)
- **Логирование**: zerolog
- **Конфигурация**: godotenv
- **Тестирование**: testify
- **Контейнеризация**: Docker

## 📂 Структура проекта
```
├── cmd/                     # Основной исполняемый файл
│   ├── main.go              # Точка входа в приложение
├── config/                  # Конфигурационные файлы
│   ├── config.go            # Логика загрузки конфигурации
├── internal/                # Внутренние модули сервиса
│   ├── adapter/             # Взаимодействие с API Allure
│   │   ├── allure-client.go # HTTP-клиент для работы с Allure API
│   │   ├── models.go        # Определение структур данных
│   ├── handler/             # HTTP-обработчики
│   │   ├── handlers.go      # Основные обработчики запросов
│   ├── service/             # Бизнес-логика
│   │   ├── allure_service.go # Allure-сервис
├── test/                    # Тесты
│   ├── client_test.go       # Тест HTTP-клиента Allure
│   ├── config_test.go       # Тест конфигурации
│   ├── handler_test.go      # Тест HTTP-обработчиков
│   ├── integration_test.go  # Интеграционные тесты
│   ├── service_test.go      # Тест сервисного слоя
├── .env                     # Файл с переменными окружения
├── .gitignore               # Файл игнорирования в Git
├── Dockerfile               # Docker-контейнеризация
├── go.mod                   # Файл зависимостей Go
├── go.sum                   # Контрольные суммы зависимостей
├── README.md                # Описание проекта
```

## ⚙️ Установка и запуск
### 🔧 Настройка переменных окружения
Перед запуском сервиса создайте файл `.env` с настройками:
```env
SERVER_PORT=8080
ALLURE_BASE_URL=https://allure.example.com
ALLURE_API_URL=/api/
ALLURE_API_TOKEN=your_api_token
ALLURE_PROJECT_ID=your_project_id
```

### 🏃‍♂️ Локальный запуск
```sh
go run cmd/main.go
```

### 🐳 Запуск в Docker
```sh
docker build -t allure-service .
docker run -p 8080:8080 --env-file .env allure-service
```

## 🛠 Тестирование
### ✅ Запуск юнит-тестов
```sh
go test ./test/... -coverpkg=./... -coverprofile=./coverage/coverage.out
```

### 📊 Анализ покрытия кода тестами
```sh
go tool cover -func=./coverage/coverage.out
```

### 🏆 Генерация HTML-отчета
```sh
go tool cover -html=./coverage/coverage.out -o ./coverage/report.html
```

## 📄 API эндпоинты
| Метод  | URL                          | Описание                               |
|--------|------------------------------|----------------------------------------|
| GET    | `/next-launch?after=<date>`  | Получение следующего запуска тестов   |
| POST   | `/export/pdf/:id`            | Генерация PDF-отчета по тесту         |
| GET    | `/export/pdf/download/:id`   | Скачивание PDF-отчета                 |

## ✨ Авторы
- **Виктория Пилипейко** — Разработка и проектирование сервиса

