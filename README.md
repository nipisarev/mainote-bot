# Mainote Bot

Telegram бот для сохранения заметок в Notion с поддержкой голосовых сообщений и утренних уведомлений.

## Возможности

- 📝 Сохранение текстовых заметок в Notion
- 🎤 Распознавание голосовых сообщений через Whisper
- 📊 Категоризация заметок (идея, задача, личное)
- 🌅 Ежедневные утренние уведомления с активными заметками

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/yourusername/mainote-bot.git
cd mainote-bot
```

2. Установите пакет:
```bash
# Установка в режиме разработки
pip install -e .

# Или установка из репозитория
pip install git+https://github.com/yourusername/mainote-bot.git
```

3. Создайте файл `.env` на основе `config.py` и заполните необходимые переменные окружения:
```
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
NOTION_API_KEY=your_notion_api_key
NOTION_DATABASE_ID=your_notion_database_id
OPENAI_API_KEY=your_openai_api_key
WEBHOOK_URL=your_webhook_url
MORNING_NOTIFICATION_TIME=08:00
NOTIFICATION_CHAT_IDS=your_chat_id1,your_chat_id2
ENABLE_MORNING_NOTIFICATIONS=true
```

## Настройка Notion

1. Создайте новую интеграцию в [Notion Developers](https://www.notion.so/my-integrations)
2. Создайте новую базу данных в Notion со следующими свойствами:
   - Name (title)
   - Type (select: idea, task, personal)
   - Status (select: active, done)
   - Source (rich text)
   - Created (date)
   - Content (rich text)
3. Предоставьте доступ к базе данных для вашей интеграции
4. Скопируйте ID базы данных из URL

## Архитектура проекта

Проект состоит из двух основных сервисов:

### 🐍 Python Bot (mainote_bot/)
Telegram бот с интеграцией Notion:

```
mainote_bot/
├── bot/                  # Обработчики команд и сообщений Telegram
│   ├── callbacks.py      # Обработчики callback-запросов
│   ├── commands.py       # Обработчики команд
│   └── messages.py       # Обработчики сообщений
├── notion/               # Интеграция с Notion
│   ├── client.py         # Клиент Notion
│   └── tasks.py          # Функции для работы с задачами
├── scheduler/            # Планировщик уведомлений
│   ├── notifications.py  # Функции отправки уведомлений
│   └── time_utils.py     # Утилиты для работы со временем
├── webhook/              # Обработка webhook-запросов
│   ├── routes.py         # FastAPI маршруты
│   └── setup.py          # Настройка webhook
├── utils/                # Утилиты
│   └── logging.py        # Настройка логирования
├── config.py             # Конфигурация приложения
├── database.py           # Операции с базой данных
├── user_preferences.py   # Управление пользовательскими настройками
└── main.py               # Основная точка входа
```

### 🐹 Go Backend (mainote_server/)
HTTP API сервис с чистой архитектурой:

```
mainote_server/
├── cmd/server/           # Точка входа приложения
├── internal/             # Приватный код приложения
│   ├── config/          # Управление конфигурацией
│   ├── delivery/http/   # HTTP обработчики и middleware
│   ├── domain/          # Бизнес-сущности и правила
│   └── usecase/         # Бизнес-логика приложения
├── go.mod               # Go модуль
└── go.sum               # Зависимости
```

### 🚀 Развертывание

**Локальная разработка:**
- Docker Compose с SQLite
- Автоматическая перезагрузка кода
- Изолированная среда

**Production (Fly.io):**
- Multi-stage Docker сборка
- Оба сервиса в одном контейнере
- Управление процессами через supervisord
- PostgreSQL для продакшена

## Запуск

### Разработка с Docker (Рекомендуется)

Для локальной разработки используйте Docker Compose:

```bash
# Быстрый старт
./dev-docker.sh start

# Просмотр логов
./dev-docker.sh logs

# Остановка
./dev-docker.sh stop
```

**Доступные команды:**
- `start` - Запустить все сервисы
- `stop` - Остановить все сервисы
- `logs` - Показать логи всех сервисов
- `logs-bot` - Показать логи только Python бота
- `logs-go` - Показать логи только Go backend
- `shell-bot` - Открыть shell в контейнере Python бота
- `build` - Собрать все сервисы
- `clean` - Остановить и удалить все контейнеры

**Сервисы будут доступны на:**
- **Python Bot**: http://localhost:8080
- **Go Backend**: http://localhost:8081
- **Health Check**: http://localhost:8081/health

**Особенности:**
- ✅ Автоматическая перезагрузка при изменении кода
- ✅ SQLite база данных (файл `./data/mainote.db`)
- ✅ Изолированная среда разработки
- ✅ Проверки здоровья сервисов

### Запуск без Docker

После установки пакета вы можете запустить бота одним из следующих способов:

```bash
# Используя скрипт run.py
python run.py

# Используя установленный пакет
mainote-bot

# Или напрямую через модуль
python -m mainote_bot.main
```

## Развертывание на fly.io

Проект использует **dual-service deployment** - оба сервиса работают в одном контейнере:

```
┌─────────────────────────────────────────┐
│             Fly.io Container            │
│                                         │
│  ┌─────────────┐    ┌─────────────────┐ │
│  │ Python Bot  │    │   Go Backend    │ │
│  │   (FastAPI) │    │  (HTTP Server)  │ │
│  │  Port 8080  │    │   Port 8081     │ │
│  └─────────────┘    └─────────────────┘ │
│           │                   │         │
│  ┌─────────────────────────────────────┐ │
│  │         Supervisord                 │ │
│  │    (Process Manager)                │ │
│  └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### Настройка и развертывание:

1. **Установите flyctl** и авторизуйтесь:
```bash
curl -L https://fly.io/install.sh | sh
flyctl auth login
```

2. **Создайте приложение:**
```bash
flyctl launch
```

3. **Настройте переменные окружения:**
```bash
flyctl secrets set TELEGRAM_BOT_TOKEN=your_token
flyctl secrets set NOTION_API_KEY=your_key
flyctl secrets set NOTION_DATABASE_ID=your_id
flyctl secrets set OPENAI_API_KEY=your_key
flyctl secrets set WEBHOOK_URL=your_url
flyctl secrets set MORNING_NOTIFICATION_TIME=08:00
flyctl secrets set NOTIFICATION_CHAT_IDS=your_chat_id1,your_chat_id2
flyctl secrets set ENABLE_MORNING_NOTIFICATIONS=true
flyctl secrets set SENTRY_DSN=your_sentry_dsn
```

4. **Разверните приложение:**
```bash
flyctl deploy
```

### API Endpoints:

**Python Bot (Port 8080 - внешний):**
- Webhook endpoints для Telegram Bot API
- FastAPI автоматическая документация

**Go Backend (Port 8081 - внутренний):**
- `GET /health` - Проверка здоровья сервиса

### Мониторинг:

```bash
# Просмотр логов
flyctl logs

# Проверка статуса
flyctl status

# Проверка здоровья Go backend (внутри контейнера)
flyctl ssh console
curl localhost:8081/health
```

## Использование

1. Найдите бота в Telegram по имени
2. Отправьте текстовое сообщение или голосовую заметку
3. Выберите категорию заметки с помощью кнопок
4. Получайте ежедневные утренние уведомления в 8:00

## Утренние уведомления

Бот может отправлять ежедневные утренние уведомления со списком активных задач из вашей базы данных Notion.

### Настройка

1. Получите ваш Chat ID в Telegram:
   - Отправьте сообщение боту [@userinfobot](https://t.me/userinfobot)
   - Он ответит вам вашим Chat ID (например, `123456789`)

2. Добавьте ваш Chat ID в переменную окружения `NOTIFICATION_CHAT_IDS`:
   - Для локального запуска: добавьте в файл `.env`
   - Для fly.io: `flyctl secrets set NOTIFICATION_CHAT_IDS=your_chat_id`

3. Настройте время уведомлений в переменной `MORNING_NOTIFICATION_TIME` (формат: `HH:MM`) или используйте команду `/settime` для настройки времени уведомлений через бота

### Команды

- `/morning` - получить утренний план на день с активными задачами
- `/start` - начать работу с ботом
- `/help` - показать справку по командам
- `/settime` - настроить время утренних уведомлений (можно указать время напрямую: `/settime 14:30`)
- `/settimezone` - настроить часовой пояс для корректной работы уведомлений

### Часовые пояса

Бот поддерживает настройку часового пояса для корректной работы уведомлений. Это особенно важно, если сервер бота работает в другом часовом поясе.

1. Используйте команду `/settimezone` для выбора вашего часового пояса
2. Выберите ваш часовой пояс из списка
3. После этого уведомления будут приходить точно в указанное вами время, с учетом вашего часового пояса

## Разработка

### Docker Compose vs Bash скрипты

Новая Docker среда заменяет старые bash скрипты (`dev-start.sh`, `dev-test-dual.sh`) и предоставляет:

| Функция | Старые скрипты | Docker Compose |
|---------|---------------|----------------|
| **База данных** | PostgreSQL/Смешанная | SQLite (простая и консистентная) |
| **Зависимости** | Ручная установка | Контейнеризированы |
| **Live Reload** | Ручной перезапуск | Автоматический |
| **Изоляция** | Система хоста | Изолированные контейнеры |
| **Очистка** | Убийство процессов | `docker-compose down` |

### Устранение неполадок

**Сервисы не запускаются:**
```bash
# Проверить статус
./dev-docker.sh status

# Проверить логи на ошибки
./dev-docker.sh logs
```

**Проблемы с базой данных:**
```bash
# Сбросить базу данных (удалит все данные!)
rm -f data/mainote.db
./dev-docker.sh restart
```

**Проблемы с контейнерами:**
```bash
# Очистить все и начать заново
./dev-docker.sh clean
./dev-docker.sh build
./dev-docker.sh start
```

**Изменения кода не отражаются:**
- **Python**: Изменения должны отражаться автоматически
- **Go**: Air должен пересобирать автоматически
- Если не работает: `./dev-docker.sh restart`
