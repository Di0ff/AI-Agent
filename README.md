# AI-Agent

AI-агент для автономного управления веб-браузером через CLI.

## Требования

- Go 1.24+
- Docker и Docker Compose (для PostgreSQL)
- Node.js и npm (для установки браузеров Playwright)
- OpenAI API ключ

## Установка

### 1. Клонирование и настройка

```bash
git clone <repository-url>
cd AI-Agent
```

### 2. Настройка окружения

Скопируйте `.env.example` в `.env` и заполните значения:

```bash
cp .env.example .env
```

Обязательно укажите:
- `OPENAI_API_KEY` - ваш ключ OpenAI API
- `DB_PASS` - пароль для PostgreSQL

### 3. Запуск PostgreSQL

```bash
docker-compose up -d postgres
```

Проверьте что БД запущена:
```bash
docker-compose ps
```

### 4. Установка зависимостей Go

```bash
go mod download
```

### 5. Установка браузеров Playwright

Для локального запуска с видимым браузером:

```bash
npm install -g playwright
npx playwright install chromium
```

## Запуск

### Локальный запуск (рекомендуется)

```bash
go run cmd/main.go
```

Или соберите бинарник:

```bash
go build -o ai-agent ./cmd
./ai-agent
```

### Команды CLI

После запуска доступны команды:

- `task <текст>` - создать новую задачу
- `tasks` - список всех задач
- `status <id>` - статус задачи
- `test-llm <задача>` - протестировать LLM
- `open <url>` - открыть URL в браузере
- `exit` - выход

## Структура проекта

```
.
├── cmd/
│   └── main.go          # Точка входа
├── internal/
│   ├── browser/         # Управление браузером (Playwright)
│   ├── cli/             # CLI интерфейс
│   ├── config/          # Конфигурация
│   ├── database/        # Работа с БД (GORM)
│   ├── llm/             # Интеграция с OpenAI
│   ├── logger/          # Логирование (Zap)
│   └── migrations/      # Миграции БД
├── docker-compose.yml   # PostgreSQL в Docker
└── .env                 # Конфигурация (не в git)
```

## Конфигурация

Все настройки в файле `.env`:

- **База данных**: `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASS`
- **OpenAI**: `OPENAI_API_KEY`, `OPENAI_MODEL`, `OPENAI_MAX_TOKENS`
- **Браузер**: `PW_HEADLESS`, `PW_USER_DATA_DIR`, `DISPLAY`
- **Логирование**: `ENV`, `LOG_LEVEL`

## Разработка

### Запуск только PostgreSQL

```bash
docker-compose up -d postgres
```

### Локальный запуск приложения

```bash
go run cmd/main.go
```

### Сборка

```bash
go build -o ai-agent ./cmd
```

## Troubleshooting

### Ошибка подключения к БД

Убедитесь что PostgreSQL запущен:
```bash
docker-compose ps
docker-compose logs postgres
```

### Браузер не запускается

Установите браузеры Playwright:
```bash
npx playwright install chromium
```

### Ошибки миграций

Проверьте путь к миграциям в `.env`:
```
MIGRATIONS_PATH=file://internal/migrations/scripts
```

