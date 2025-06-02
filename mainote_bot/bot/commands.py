from telegram import Bot, Update, InlineKeyboardButton, InlineKeyboardMarkup, KeyboardButton, ReplyKeyboardMarkup
from telegram.ext import ContextTypes
from mainote_bot.config import TELEGRAM_BOT_TOKEN, NOTIFICATION_CHAT_IDS
from mainote_bot.utils.logging import logger
from mainote_bot.notion.tasks import get_active_tasks, format_morning_notification
import mainote_bot.user_preferences as user_preferences
import pytz
from mainote_bot.scheduler.notifications import force_notification_recalculation
from timezonefinder import TimezoneFinder
from datetime import datetime

async def start_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /start command."""
    try:
        chat_id = update.effective_chat.id
        
        # First try to get timezone from user's location if available
        if update.effective_message and update.effective_message.location:
            await handle_location(update, context)
            return
            
        # If no location, ask user to share it
        keyboard = [[KeyboardButton("📍 Поделиться местоположением", request_location=True)]]
        reply_markup = ReplyKeyboardMarkup(keyboard, one_time_keyboard=True)
        
        # Send location request message
        await context.bot.send_message(
            chat_id=chat_id,
            text="👋 Привет! Я бот для сохранения заметок в Notion.\n\n"
                 "Для корректной работы уведомлений, пожалуйста, поделитесь вашим местоположением.\n"
                 "Это поможет установить правильный часовой пояс.",
            reply_markup=reply_markup
        )
        
        # Send main welcome message
        await context.bot.send_message(
            chat_id=chat_id,
            text="Просто отправь мне текст, и я сохраню его как заметку.\n"
                 "После сохранения ты сможешь выбрать тип заметки:\n"
                 "💡 Идея\n"
                 "✅ Задача\n"
                 "🏖 Личное"
        )
        
        logger.info(f"Sent welcome messages to {chat_id}")
    except Exception as e:
        logger.error(f"Error in start command: {str(e)}", exc_info=True)

async def handle_location(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle location message to set user's timezone."""
    try:
        chat_id = update.effective_chat.id
        location = update.effective_message.location
        
        if location:
            # Use timezonefinder to get timezone from coordinates
            tf = TimezoneFinder()
            timezone_str = tf.timezone_at(lat=location.latitude, lng=location.longitude)
            
            if timezone_str:
                # Save user's timezone
                if await user_preferences.set_user_timezone(chat_id, timezone_str):
                    logger.info(f"Set timezone to {timezone_str} for chat ID {chat_id}")
                    force_notification_recalculation()
                    
                    # Get timezone offset for display
                    tz = pytz.timezone(timezone_str)
                    offset = tz.utcoffset(datetime.now()).total_seconds() / 3600
                    offset_str = f"UTC{'+' if offset >= 0 else ''}{int(offset)}"
                    
                    # Send timezone confirmation
                    await context.bot.send_message(
                        chat_id=chat_id,
                        text=f"🌐 Часовой пояс установлен на {timezone_str} ({offset_str}).\n\n"
                             f"Теперь уведомления будут приходить в соответствии с вашим часовым поясом."
                    )
                    
                    # Send main welcome message
                    await context.bot.send_message(
                        chat_id=chat_id,
                        text="Просто отправь мне текст, и я сохраню его как заметку.\n"
                             "После сохранения ты сможешь выбрать тип заметки:\n"
                             "💡 Идея\n"
                             "✅ Задача\n"
                             "🏖 Личное"
                    )
                else:
                    logger.error(f"Failed to save timezone {timezone_str} for chat ID {chat_id}")
                    await context.bot.send_message(
                        chat_id=chat_id,
                        text="Произошла ошибка при сохранении часового пояса. Попробуйте позже."
                    )
            else:
                logger.error(f"Could not determine timezone for coordinates: {location.latitude}, {location.longitude}")
                await context.bot.send_message(
                    chat_id=chat_id,
                    text="Не удалось определить часовой пояс по вашему местоположению. Попробуйте использовать команду /settimezone"
                )
    except Exception as e:
        logger.error(f"Error handling location: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=chat_id,
            text="Произошла ошибка при обработке местоположения. Попробуйте позже."
        )

async def help_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /help command."""
    try:
        chat_id = update.effective_chat.id
        await context.bot.send_message(
            chat_id=chat_id,
            text="📝 Как пользоваться ботом:\n\n"
                 "1. Отправь любой текст - он будет сохранен как заметка\n"
                 "2. После сохранения выбери тип заметки\n"
                 "3. Готово! Заметка появится в твоем Notion\n\n"
                 "Команды:\n"
                 "/start - начать работу с ботом\n"
                 "/help - показать это сообщение\n"
                 "/morning - получить утренний план на день\n"
                 "/settime - настроить время утренних уведомлений (можно указать время напрямую: /settime 14:30)\n"
                 
        )
        logger.info(f"Sent help message to {chat_id}")
    except Exception as e:
        logger.error(f"Error in help command: {str(e)}", exc_info=True)

async def morning_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /morning command to manually trigger morning notification."""
    try:
        chat_id = update.effective_chat.id

        # Add this chat_id to notification recipients if not already there
        chat_id_str = str(chat_id)
        if chat_id_str not in NOTIFICATION_CHAT_IDS and NOTIFICATION_CHAT_IDS != ['']:
            logger.info(f"Adding chat ID {chat_id} to notification recipients")
            # Note: This won't persist after restart, would need to save to a file or database

        # Get active tasks
        tasks = await get_active_tasks()

        # Format notification message
        message = await format_morning_notification(tasks)

        # Send the notification
        await context.bot.send_message(chat_id=chat_id, text=message)
        logger.info(f"Sent manual morning notification to {chat_id}")
    except Exception as e:
        logger.error(f"Error in morning command: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Произошла ошибка при получении утреннего плана. Попробуйте позже."
        )

async def settimezone_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /settimezone command to configure user timezone."""
    try:
        chat_id = update.effective_chat.id

        # Get current timezone for this user
        current_timezone = user_preferences.get_user_timezone(chat_id)
        current_timezone_msg = f"Текущий часовой пояс: {current_timezone}" if current_timezone else "Часовой пояс не настроен"

        # Common timezones for Russia and nearby regions
        common_timezones = [
            ("Europe/Moscow", "Москва (UTC+3)"),
            ("Europe/Kaliningrad", "Калининград (UTC+2)"),
            ("Europe/Samara", "Самара (UTC+4)"),
            ("Asia/Yekaterinburg", "Екатеринбург (UTC+5)"),
            ("Asia/Omsk", "Омск (UTC+6)"),
            ("Asia/Krasnoyarsk", "Красноярск (UTC+7)"),
            ("Asia/Irkutsk", "Иркутск (UTC+8)"),
            ("Asia/Yakutsk", "Якутск (UTC+9)"),
            ("Asia/Vladivostok", "Владивосток (UTC+10)"),
            ("Asia/Magadan", "Магадан (UTC+11)"),
            ("Asia/Kamchatka", "Камчатка (UTC+12)")
        ]

        # Create buttons for timezone selection
        keyboard = []
        for tz_name, tz_display in common_timezones:
            keyboard.append([InlineKeyboardButton(tz_display, callback_data=f"timezone:{tz_name}")])

        reply_markup = InlineKeyboardMarkup(keyboard)

        # Create a new bot instance for this event loop
        local_bot = Bot(token=TELEGRAM_BOT_TOKEN)

        await local_bot.send_message(
            chat_id=chat_id,
            text=f"🌐 Выберите ваш часовой пояс для корректной работы уведомлений:\n\n{current_timezone_msg}",
            reply_markup=reply_markup
        )
        logger.info(f"Sent timezone selection buttons to {chat_id}")
    except Exception as e:
        logger.error(f"Error in settimezone command: {str(e)}", exc_info=True)
        # Create a new bot instance for this event loop
        local_bot = Bot(token=TELEGRAM_BOT_TOKEN)
        await local_bot.send_message(
            chat_id=update.effective_chat.id,
            text="Произошла ошибка при настройке часового пояса. Попробуйте позже."
        )

async def settime_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /settime command to configure notification time."""
    try:
        chat_id = update.effective_chat.id

        # Check if user has timezone set
        user_timezone = await user_preferences.get_user_timezone(chat_id)
        if not user_timezone:
            # Create a new bot instance for this event loop
            local_bot = Bot(token=TELEGRAM_BOT_TOKEN)
            await local_bot.send_message(
                chat_id=chat_id,
                text="⚠️ Для корректной работы уведомлений необходимо установить часовой пояс.\n\n"
                     "Пожалуйста, используйте команду /settimezone для выбора вашего часового пояса."
            )
            return

        # Check if a time argument was provided
        if context.args and len(context.args) > 0:
            time_arg = context.args[0]

            # Validate time format (HH:MM)
            try:
                hour, minute = map(int, time_arg.split(':'))
                if 0 <= hour < 24 and 0 <= minute < 60:
                    # Save user preference
                    if await user_preferences.set_user_notification_time(chat_id, time_arg):
                        logger.info(f"Set notification time to {time_arg} for chat ID {chat_id}")

                        # Create a new bot instance for this event loop
                        local_bot = Bot(token=TELEGRAM_BOT_TOKEN)

                        # Get timezone offset for display
                        tz = pytz.timezone(user_timezone)
                        offset = tz.utcoffset(datetime.now()).total_seconds() / 3600
                        offset_str = f"UTC{'+' if offset >= 0 else ''}{int(offset)}"

                        # Send confirmation
                        await local_bot.send_message(
                            chat_id=chat_id,
                            text=f"⏰ Время утренних уведомлений установлено на {time_arg} ({user_timezone}, {offset_str}).\n\n"
                                 f"Вы будете получать уведомления ежедневно в {time_arg} по вашему местному времени."
                        )
                        return
                    else:
                        logger.error(f"Failed to save notification time for chat ID {chat_id}")
                else:
                    logger.error(f"Invalid time format: {time_arg}")
            except Exception as e:
                logger.error(f"Error parsing time argument: {str(e)}", exc_info=True)
                # Continue to show time selection buttons

        # Get current notification time for this user
        current_time = await user_preferences.get_user_notification_time(chat_id)
        current_time_msg = f"Текущее время уведомлений: {current_time}" if current_time else "Время уведомлений не настроено"

        # Create buttons for common times (more options)
        keyboard = [
            [
                InlineKeyboardButton("06:00", callback_data="time:06:00"),
                InlineKeyboardButton("07:00", callback_data="time:07:00"),
                InlineKeyboardButton("08:00", callback_data="time:08:00")
            ],
            [
                InlineKeyboardButton("09:00", callback_data="time:09:00"),
                InlineKeyboardButton("10:00", callback_data="time:10:00"),
                InlineKeyboardButton("11:00", callback_data="time:11:00")
            ],
            [
                InlineKeyboardButton("12:00", callback_data="time:12:00"),
                InlineKeyboardButton("14:00", callback_data="time:14:00"),
                InlineKeyboardButton("16:00", callback_data="time:16:00")
            ],
            [
                InlineKeyboardButton("18:00", callback_data="time:18:00"),
                InlineKeyboardButton("20:00", callback_data="time:20:00"),
                InlineKeyboardButton("22:00", callback_data="time:22:00")
            ],
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)

        # Create a new bot instance for this event loop
        local_bot = Bot(token=TELEGRAM_BOT_TOKEN)

        await local_bot.send_message(
            chat_id=chat_id,
            text=f"⏰ Выберите время для утренних уведомлений или используйте команду '/settime ЧЧ:ММ' для установки произвольного времени:\n\n{current_time_msg}",
            reply_markup=reply_markup
        )
        logger.info(f"Sent time selection buttons to {chat_id}")
    except Exception as e:
        logger.error(f"Error in settime command: {str(e)}", exc_info=True)
        # Create a new bot instance for this event loop
        local_bot = Bot(token=TELEGRAM_BOT_TOKEN)
        await local_bot.send_message(
            chat_id=update.effective_chat.id,
            text="Произошла ошибка при настройке времени уведомлений. Попробуйте позже."
        )