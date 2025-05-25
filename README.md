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

## Структура проекта

Проект организован в виде Python-пакета с разделением функциональности по модулям:

```
mainote_bot/
├── __init__.py
├── bot/                  # Обработчики команд и сообщений Telegram
│   ├── __init__.py
│   ├── callbacks.py      # Обработчики callback-запросов
│   ├── commands.py       # Обработчики команд
│   └── messages.py       # Обработчики сообщений
├── config.py             # Конфигурация приложения
├── main.py               # Основная точка входа
├── notion/               # Интеграция с Notion
│   ├── __init__.py
│   ├── client.py         # Клиент Notion
│   └── tasks.py          # Функции для работы с задачами
├── scheduler/            # Планировщик уведомлений
│   ├── __init__.py
│   ├── notifications.py  # Функции отправки уведомлений
│   └── time_utils.py     # Утилиты для работы со временем
├── user_preferences.py   # Управление пользовательскими настройками
├── utils/                # Утилиты
│   ├── __init__.py
│   └── logging.py        # Настройка логирования
└── webhook/              # Обработка webhook-запросов
    ├── __init__.py
    ├── routes.py         # Маршруты Flask
    └── setup.py          # Настройка webhook
```

Дополнительные файлы:
- `setup.py` - Файл для установки пакета
- `run.py` - Скрипт для запуска бота
- `requirements.txt` - Зависимости проекта
- `Dockerfile` - Файл для создания Docker-образа
- `.dockerignore` - Файлы, исключаемые из Docker-образа
- `fly.toml` - Конфигурация для развертывания на fly.io

## Запуск

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

1. Установите [flyctl](https://fly.io/docs/hands-on/install-flyctl/)
2. Авторизуйтесь:
```bash
flyctl auth login
```

3. Создайте приложение:
```bash
flyctl launch
```

4. Настройте переменные окружения:
```bash
flyctl secrets set TELEGRAM_BOT_TOKEN=your_token
flyctl secrets set NOTION_API_KEY=your_key
flyctl secrets set NOTION_DATABASE_ID=your_id
flyctl secrets set OPENAI_API_KEY=your_key
flyctl secrets set WEBHOOK_URL=your_url
flyctl secrets set MORNING_NOTIFICATION_TIME=08:00
flyctl secrets set NOTIFICATION_CHAT_IDS=your_chat_id1,your_chat_id2
flyctl secrets set ENABLE_MORNING_NOTIFICATIONS=true
```

5. Разверните приложение:
```bash
flyctl deploy
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
