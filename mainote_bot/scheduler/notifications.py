import asyncio
from datetime import datetime
import pytz
from telegram import Bot
from mainote_bot.config import (
    TELEGRAM_BOT_TOKEN, NOTIFICATION_CHAT_IDS, 
    MORNING_NOTIFICATION_TIME, ENABLE_MORNING_NOTIFICATIONS
)
from mainote_bot.utils.logging import logger
from mainote_bot.notion.tasks import get_active_tasks, format_morning_notification
import mainote_bot.user_preferences as user_preferences
from mainote_bot.scheduler.time_utils import (
    calculate_user_notification_time, calculate_default_notification_times,
    find_earliest_notification
)

async def send_notifications_to_users(users_to_notify):
    """Send notifications to all users scheduled for a specific time."""
    logger.info(f"Preparing to send notifications to {len(users_to_notify)} users: {users_to_notify}")

    # Create a new bot instance for this event loop
    local_bot = Bot(token=TELEGRAM_BOT_TOKEN)
    logger.info("Created new bot instance for sending notifications")

    # Get active tasks
    tasks = await get_active_tasks()
    logger.info(f"Found {len(tasks)} active tasks for notifications")

    # Format notification message
    message = await format_morning_notification(tasks)
    logger.info("Formatted notification message")

    # Send to all users scheduled for this time
    notification_sent_count = 0
    for user_id in users_to_notify:
        try:
            chat_id = int(user_id.strip())
            logger.info(f"Attempting to send notification to chat ID {chat_id}...")
            await local_bot.send_message(chat_id=chat_id, text=message)
            logger.info(f"Successfully sent morning notification to chat ID {chat_id}")
            notification_sent_count += 1
        except Exception as e:
            logger.error(f"Error sending notification to chat ID {user_id}: {str(e)}", exc_info=True)

    logger.info(f"Sent notifications to {notification_sent_count} out of {len(users_to_notify)} users")

async def send_morning_notification():
    """Send morning notification with active tasks to all registered users."""
    if not ENABLE_MORNING_NOTIFICATIONS:
        logger.info("Morning notifications are disabled")
        return

    if not NOTIFICATION_CHAT_IDS or NOTIFICATION_CHAT_IDS == ['']:
        logger.warning("No chat IDs configured for morning notifications")
        return

    try:
        # Get active tasks
        tasks = await get_active_tasks()

        # Format notification message
        message = await format_morning_notification(tasks)

        # Create a new bot instance for this event loop
        local_bot = Bot(token=TELEGRAM_BOT_TOKEN)

        # Send to all configured chat IDs
        for chat_id in NOTIFICATION_CHAT_IDS:
            if not chat_id:
                continue

            try:
                chat_id = int(chat_id.strip())
                await local_bot.send_message(chat_id=chat_id, text=message)
                logger.info(f"Sent morning notification to chat ID {chat_id}")
            except Exception as e:
                logger.error(f"Error sending notification to chat ID {chat_id}: {str(e)}", exc_info=True)

    except Exception as e:
        logger.error(f"Error in send_morning_notification: {str(e)}", exc_info=True)

# Global variable to track if scheduler is running
scheduler_running = False
# Global variable to track the last preference check time
last_preference_check = None
# Global variable to force recalculation
force_recalculation = False

def force_notification_recalculation():
    """Force the scheduler to recalculate notification times."""
    global force_recalculation
    force_recalculation = True
    logger.info("Forcing notification time recalculation")

async def schedule_morning_notifications():
    """Schedule the morning notifications to run daily."""
    global last_preference_check, force_recalculation
    logger.info("🔄 Starting morning notification scheduler loop")
    
    while True:
        try:
            # Get current time in UTC
            now = datetime.now(pytz.UTC)
            
            # Check if we need to recalculate (every 5 minutes or if forced)
            should_recalculate = (
                last_preference_check is None or
                (now - last_preference_check).total_seconds() >= 300 or  # 5 minutes
                force_recalculation
            )
            
            if should_recalculate:
                logger.info(f"Recalculating notification times (forced: {force_recalculation})")
                last_preference_check = now
                force_recalculation = False
                
                # Get all users with preferences
                users_with_prefs = await user_preferences.get_all_users_with_preferences()
                logger.info(f"Found {len(users_with_prefs)} users with preferences: {list(users_with_prefs)}")

                # Dictionary to store next notification time for each user
                next_notifications = {}

                # Calculate next notification time for each user with custom preferences
                for user_id in users_with_prefs:
                    user_time = await user_preferences.get_user_notification_time(user_id)
                    user_timezone_str = await user_preferences.get_user_timezone(user_id)

                    if user_time:
                        notification_info = await calculate_user_notification_time(user_id, user_time, user_timezone_str, now)
                        if notification_info:
                            next_notifications[user_id] = notification_info

                # For users without custom preferences, use the default time
                default_users = [chat_id for chat_id in NOTIFICATION_CHAT_IDS if chat_id and chat_id not in users_with_prefs]
                default_notifications = calculate_default_notification_times(default_users, now, MORNING_NOTIFICATION_TIME)
                next_notifications.update(default_notifications)

                # Find the earliest notification time
                _, next_time, seconds_until_next = find_earliest_notification(next_notifications)

                if next_time:
                    logger.info(f"Next notification scheduled for {next_time.strftime('%Y-%m-%d %H:%M:%S')} UTC")
                    
                    # Wait until the next notification time or until next preference check
                    wait_time = min(seconds_until_next, 60)  # Wait at most 1 minute
                    await asyncio.sleep(wait_time)
                    
                    # If we've reached the notification time
                    if wait_time == seconds_until_next:
                        current_time = datetime.now(pytz.UTC)
                        logger.info(f"Notification time reached! Current time: {current_time.strftime('%Y-%m-%d %H:%M:%S')} UTC")

                        # Send notifications to all users scheduled for this time
                        users_to_notify = [user_id for user_id in next_notifications 
                                          if abs(next_notifications[user_id]['seconds'] - seconds_until_next) < 60]  # Within 1 minute

                        await send_notifications_to_users(users_to_notify)
                else:
                    # If no users to notify, wait for next preference check
                    logger.info("No users to notify. Checking again in 1 minute.")
                    await asyncio.sleep(60)  # 1 minute

            else:
                # If no recalculation needed, wait a bit before checking again
                await asyncio.sleep(60)  # Check every minute

        except Exception as e:
            logger.error(f"Error in schedule_morning_notifications: {str(e)}", exc_info=True)
            # Wait a bit before retrying
            await asyncio.sleep(60)

def start_scheduler():
    """Start the scheduler in the main event loop."""
    global scheduler_running
    try:
        if scheduler_running:
            logger.info("Scheduler already running")
            return

        if ENABLE_MORNING_NOTIFICATIONS:
            logger.info(
                f"Starting morning notification scheduler (time: {MORNING_NOTIFICATION_TIME}, recipients: {NOTIFICATION_CHAT_IDS})"
            )

            # Create the scheduler task
            asyncio.create_task(schedule_morning_notifications())
            scheduler_running = True
            logger.info("Scheduler task created and started")
        else:
            logger.info("Morning notifications are disabled")
    except Exception as e:
        logger.error(f"Error starting scheduler: {str(e)}", exc_info=True)