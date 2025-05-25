import os
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN = os.getenv('TELEGRAM_BOT_TOKEN')

# Notion Configuration
NOTION_API_KEY = os.getenv('NOTION_API_KEY')
NOTION_DATABASE_ID = os.getenv('NOTION_DATABASE_ID')

# OpenAI Configuration
OPENAI_API_KEY = os.getenv('OPENAI_API_KEY')

# Webhook Configuration
WEBHOOK_URL = os.getenv('WEBHOOK_URL')

# Sentry DSN Configuration
SENTRY_DSN = os.getenv('SENTRY_DSN')

# Morning Notification Configuration
MORNING_NOTIFICATION_TIME = os.getenv('MORNING_NOTIFICATION_TIME', '08:00')
NOTIFICATION_CHAT_IDS = os.getenv('NOTIFICATION_CHAT_IDS', '').split(',')
ENABLE_MORNING_NOTIFICATIONS = 'true'

# Note Categories
NOTE_CATEGORIES = {
    'idea': '💡 Идея',
    'task': '✅ Задача',
    'personal': '🏖 Личное'
}

# Database Schema
NOTION_DB_SCHEMA = {
    'Name': 'title',
    'Type': 'select',
    'Status': 'select',
    'Source': 'rich_text',
    'Created': 'date',
    'Content': 'rich_text'
}

# Constants
ERROR_PROCESSING_REQUEST = "Произошла ошибка при обработке запроса. Попробуйте позже."