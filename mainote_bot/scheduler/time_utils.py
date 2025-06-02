from datetime import timedelta
import pytz
from mainote_bot.utils.logging import logger

def calculate_time_without_timezone(user_id, hour, minute, user_time, now, is_fallback=False):
    """Calculate notification time for a user without timezone set or as fallback."""
    user_target_time = now.replace(hour=hour, minute=minute, second=0, microsecond=0)
    
    # If the target time is more than 5 minutes in the past, schedule for tomorrow
    if (now - user_target_time).total_seconds() > 60:  # 5 minutes
        user_target_time = user_target_time + timedelta(days=1)

    notification_info = {
        'time': user_target_time,
        'seconds': (user_target_time - now).total_seconds(),
        'user_time': user_time,
        'timezone': 'UTC (fallback)' if is_fallback else 'UTC (server)'
    }

    log_prefix = "Fallback: " if is_fallback else ""
    logger.info(f"{log_prefix}User {user_id} will receive notification at {user_target_time.strftime('%Y-%m-%d %H:%M:%S')} (server time)")
    return notification_info

async def calculate_time_with_timezone(user_id, hour, minute, user_timezone_str, user_time, now):
    """Calculate notification time for a user with timezone set."""
    try:
        # Get user's timezone
        user_timezone = pytz.timezone(user_timezone_str)

        # Get current time in user's timezone
        now_user_tz = now.astimezone(user_timezone)

        # Create target time in user's timezone
        user_target_time = now_user_tz.replace(hour=hour, minute=minute, second=0, microsecond=0)

        # If the target time is more than 5 minutes in the past, schedule for tomorrow
        if (now_user_tz - user_target_time).total_seconds() > 60:  # 5 minutes
            user_target_time = user_target_time + timedelta(days=1)

        # Convert back to UTC for scheduling
        utc_target_time = user_target_time.astimezone(pytz.UTC)

        notification_info = {
            'time': utc_target_time,
            'seconds': (utc_target_time - now).total_seconds(),
            'user_time': user_time,
            'timezone': user_timezone_str
        }

        logger.info(f"User {user_id} will receive notification at {user_time} in {user_timezone_str} (server time: {utc_target_time.strftime('%Y-%m-%d %H:%M:%S')})")
        return notification_info
    except Exception as e:
        logger.error(f"Error converting timezone for user {user_id}: {str(e)}", exc_info=True)
        # Fall back to server timezone if there's an error
        return calculate_time_without_timezone(user_id, hour, minute, user_time, now, is_fallback=True)

async def calculate_user_notification_time(user_id, user_time, user_timezone_str, now):
    """Calculate the next notification time for a user with custom preferences."""
    try:
        hour, minute = map(int, user_time.split(':'))

        # If user has set a timezone, use it for calculations
        if user_timezone_str:
            return await calculate_time_with_timezone(user_id, hour, minute, user_timezone_str, user_time, now)
        else:
            # If no timezone set, use server timezone
            return calculate_time_without_timezone(user_id, hour, minute, user_time, now)
    except Exception as e:
        logger.error(f"Error parsing user time preference for {user_id}: {str(e)}", exc_info=True)
        return None

def calculate_default_notification_times(default_users, now, default_time):
    """Calculate notification times for users without custom preferences."""
    next_notifications = {}

    if not default_users:
        return next_notifications

    # Parse regular notification time
    hour, minute = map(int, default_time.split(':'))

    # Calculate time until next regular notification in UTC
    default_target_time = now.replace(hour=hour, minute=minute, second=0, microsecond=0)
    if now >= default_target_time:
        # If we've already passed the target time today, schedule for tomorrow
        default_target_time = default_target_time + timedelta(days=1)

    # Add default notification time for users without custom preferences
    for user_id in default_users:
        next_notifications[user_id] = {
            'time': default_target_time,
            'seconds': (default_target_time - now).total_seconds(),
            'user_time': default_time,
            'timezone': 'UTC (default)'
        }

    logger.info(f"{len(default_users)} users will receive notification at default time {default_target_time.strftime('%Y-%m-%d %H:%M:%S')} UTC")
    return next_notifications

def find_earliest_notification(next_notifications):
    """Find the earliest notification time among all users."""
    if not next_notifications:
        return None, None, None

    next_user_id = min(next_notifications, key=lambda x, n=next_notifications: n[x]['seconds'])
    next_time = next_notifications[next_user_id]['time']
    seconds_until_next = next_notifications[next_user_id]['seconds']

    logger.info(f"Next notification in {seconds_until_next:.2f} seconds (at {next_time.strftime('%Y-%m-%d %H:%M:%S')})")
    return next_user_id, next_time, seconds_until_next