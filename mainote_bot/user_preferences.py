import json
import os
from mainote_bot.utils.logging import logger

# File to store user preferences
PREFERENCES_FILE = "user_preferences.json"

def load_preferences():
    """Load user preferences from JSON file."""
    if not os.path.exists(PREFERENCES_FILE):
        return {}

    try:
        with open(PREFERENCES_FILE, 'r') as f:
            return json.load(f)
    except Exception as e:
        logger.error(f"Error loading user preferences: {str(e)}", exc_info=True)
        return {}

def save_preferences(preferences):
    """Save user preferences to JSON file."""
    try:
        with open(PREFERENCES_FILE, 'w') as f:
            json.dump(preferences, f, indent=2)
        return True
    except Exception as e:
        logger.error(f"Error saving user preferences: {str(e)}", exc_info=True)
        return False

def get_user_notification_time(chat_id):
    """Get notification time for a specific user."""
    chat_id_str = str(chat_id)
    preferences = load_preferences()

    # Return user's preferred time or None if not set
    return preferences.get(chat_id_str, {}).get('notification_time')

def get_user_timezone(chat_id):
    """Get timezone for a specific user."""
    chat_id_str = str(chat_id)
    preferences = load_preferences()

    # Return user's timezone or None if not set
    return preferences.get(chat_id_str, {}).get('timezone')

def set_user_notification_time(chat_id, time):
    """Set notification time for a specific user."""
    chat_id_str = str(chat_id)
    preferences = load_preferences()

    # Initialize user preferences if not exists
    if chat_id_str not in preferences:
        preferences[chat_id_str] = {}

    # Set notification time
    preferences[chat_id_str]['notification_time'] = time

    # Save preferences
    return save_preferences(preferences)

def set_user_timezone(chat_id, timezone):
    """Set timezone for a specific user."""
    chat_id_str = str(chat_id)
    preferences = load_preferences()

    # Initialize user preferences if not exists
    if chat_id_str not in preferences:
        preferences[chat_id_str] = {}

    # Set timezone
    preferences[chat_id_str]['timezone'] = timezone

    # Save preferences
    return save_preferences(preferences)

def get_all_users_with_preferences():
    """Get all users who have set preferences."""
    return load_preferences().keys()