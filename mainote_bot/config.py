import os
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN = os.getenv('TELEGRAM_BOT_TOKEN')

# Server Configuration
SERVER_URL = os.getenv('MAINOTE_SERVER_URL', 'http://mainote-server:8081')

# OpenAI Configuration
OPENAI_API_KEY = os.getenv('OPENAI_API_KEY')

# Webhook Configuration
WEBHOOK_URL = os.getenv('WEBHOOK_URL')

# Internal API Configuration
INTERNAL_API_KEY = os.getenv('INTERNAL_API_KEY')

# Sentry DSN Configuration
SENTRY_DSN = os.getenv('SENTRY_DSN')

# Morning Notification Configuration
MORNING_NOTIFICATION_TIME = os.getenv('MORNING_NOTIFICATION_TIME', '08:00')
NOTIFICATION_CHAT_IDS = os.getenv('NOTIFICATION_CHAT_IDS', '').split(',')
# Morning notifications can be toggled via environment variable
# Expect values like "true"/"false" (case-insensitive)
ENABLE_MORNING_NOTIFICATIONS = os.getenv('ENABLE_MORNING_NOTIFICATIONS', 'true').lower() == 'true'

# Note Categories
NOTE_CATEGORIES = {
    'idea': '💡 Idea',
    'task': '✅ Task',
    'personal': '🏖 Personal'
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
ERROR_PROCESSING_REQUEST = "An error occurred while processing the request. Please try again later."