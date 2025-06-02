import json
import os
from mainote_bot.utils.logging import logger
from mainote_bot.database import (
    get_user_preferences as db_get_user_preferences,
    set_user_preferences as db_set_user_preferences,
    get_all_users_with_preferences as db_get_all_users_with_preferences
)

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

async def get_user_notification_time(chat_id):
    """Get notification time for a specific user."""
    preferences = await db_get_user_preferences(chat_id)
    return preferences.get('notification_time')

async def get_user_timezone(chat_id):
    """Get timezone for a specific user."""
    preferences = await db_get_user_preferences(chat_id)
    return preferences.get('timezone')

async def set_user_notification_time(chat_id, time):
    """Set notification time for a specific user."""
    try:
        # Get existing preferences or create new ones
        preferences = await db_get_user_preferences(chat_id)
        if not preferences:
            preferences = {'notification_time': None, 'timezone': None}
        
        # Update notification time
        preferences['notification_time'] = time
        
        # Save to database
        success = await db_set_user_preferences(chat_id, preferences)
        if not success:
            logger.error(f"Failed to save notification time {time} for chat ID {chat_id}")
        return success
    except Exception as e:
        logger.error(f"Error setting notification time for chat ID {chat_id}: {str(e)}", exc_info=True)
        return False

async def set_user_timezone(chat_id, timezone):
    """Set timezone for a specific user."""
    try:
        # Get existing preferences or create new ones
        preferences = await db_get_user_preferences(chat_id)
        if not preferences:
            preferences = {'notification_time': None, 'timezone': None}
        
        # Update timezone
        preferences['timezone'] = timezone
        
        # Save to database
        success = await db_set_user_preferences(chat_id, preferences)
        if not success:
            logger.error(f"Failed to save timezone {timezone} for chat ID {chat_id}")
        return success
    except Exception as e:
        logger.error(f"Error setting timezone for chat ID {chat_id}: {str(e)}", exc_info=True)
        return False

async def get_all_users_with_preferences():
    """Get all users who have set preferences."""
    return await db_get_all_users_with_preferences()