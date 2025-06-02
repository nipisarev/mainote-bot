from telegram import Bot, Update
from telegram.ext import ContextTypes
from mainote_bot.config import TELEGRAM_BOT_TOKEN, ERROR_PROCESSING_REQUEST
from mainote_bot.utils.logging import logger
from mainote_bot.notion.tasks import update_note_type
import mainote_bot.user_preferences as user_preferences
import pytz
from mainote_bot.scheduler.notifications import force_notification_recalculation
from datetime import datetime

async def send_callback_response(bot, query, text):
    """Send a response to a callback query."""
    try:
        await bot.edit_message_text(
            chat_id=query.message.chat_id,
            message_id=query.message.message_id,
            text=text
        )
    except Exception as e:
        logger.error(f"Error sending callback response: {str(e)}", exc_info=True)

async def handle_note_type_callback(bot, query, data_value):
    """Handle note type selection callback."""
    # Further split the value for type:category:id format
    type_parts = data_value.split(":", 1)
    if len(type_parts) != 2:
        logger.error(f"Invalid type callback data format: {data_value}")
        await send_callback_response(bot, query, ERROR_PROCESSING_REQUEST)
        return

    note_type = type_parts[0]
    page_id = type_parts[1]

    # Map callback data to human-readable type names
    type_names = {
        "idea": "Идея",
        "task": "Задача",
        "personal": "Личное"
    }

    # Update the note type in Notion
    if update_note_type(page_id, note_type):
        # Send confirmation
        await send_callback_response(
            bot, 
            query, 
            f"Заметка сохранена как {type_names.get(note_type, note_type)} 👍"
        )
    else:
        await send_callback_response(bot, query, ERROR_PROCESSING_REQUEST)

async def handle_time_callback(bot, query, time_value):
    """Handle notification time selection callback."""
    chat_id = query.message.chat_id

    # Check if user has timezone set
    user_timezone = await user_preferences.get_user_timezone(chat_id)
    if not user_timezone:
        await send_callback_response(
            bot, 
            query, 
            "⚠️ Для корректной работы уведомлений необходимо установить часовой пояс.\n\n"
            "Пожалуйста, используйте команду /settimezone для выбора вашего часового пояса."
        )
        return

    # Validate time format (HH:MM)
    try:
        hour, minute = map(int, time_value.split(':'))
        if 0 <= hour < 24 and 0 <= minute < 60:
            # Save user preference
            if await user_preferences.set_user_notification_time(chat_id, time_value):
                logger.info(f"Set notification time to {time_value} for chat ID {chat_id}")
                
                # Force notification recalculation
                force_notification_recalculation()

                # Get timezone offset for display
                tz = pytz.timezone(user_timezone)
                offset = tz.utcoffset(datetime.now()).total_seconds() / 3600
                offset_str = f"UTC{'+' if offset >= 0 else ''}{int(offset)}"

                # Send confirmation
                await send_callback_response(
                    bot, 
                    query, 
                    f"⏰ Время утренних уведомлений установлено на {time_value} ({user_timezone}, {offset_str}).\n\n"
                    f"Вы будете получать уведомления ежедневно в {time_value} по вашему местному времени."
                )
            else:
                logger.error(f"Failed to save notification time for chat ID {chat_id}")
                await send_callback_response(
                    bot, 
                    query, 
                    "Произошла ошибка при сохранении времени уведомлений. Попробуйте позже."
                )
        else:
            logger.error(f"Invalid time format: {time_value}")
            await send_callback_response(
                bot, 
                query, 
                "Неверный формат времени. Пожалуйста, используйте формат ЧЧ:ММ (например, 08:00)."
            )
    except Exception as e:
        logger.error(f"Error parsing time: {str(e)}", exc_info=True)
        await send_callback_response(
            bot, 
            query, 
            "Неверный формат времени. Пожалуйста, используйте формат ЧЧ:ММ (например, 08:00)."
        )

async def handle_timezone_callback(bot, query, timezone_value):
    """Handle timezone selection callback."""
    chat_id = query.message.chat_id

    # Validate timezone
    try:
        # Check if timezone is valid
        pytz.timezone(timezone_value)

        # Save user preference
        if await user_preferences.set_user_timezone(chat_id, timezone_value):
            logger.info(f"Set timezone to {timezone_value} for chat ID {chat_id}")
            
            # Force notification recalculation
            force_notification_recalculation()

            # Send confirmation
            await send_callback_response(
                bot, 
                query, 
                f"🌐 Часовой пояс установлен на {timezone_value}.\n\n"
                f"Теперь уведомления будут приходить в соответствии с вашим часовым поясом."
            )
        else:
            logger.error(f"Failed to save timezone for chat ID {chat_id}")
            await send_callback_response(
                bot, 
                query, 
                "Произошла ошибка при сохранении часового пояса. Попробуйте позже."
            )
    except Exception as e:
        logger.error(f"Error setting timezone: {str(e)}", exc_info=True)
        await send_callback_response(
            bot, 
            query, 
            "Неверный формат часового пояса. Пожалуйста, выберите из предложенных вариантов."
        )

async def button_callback(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle button clicks from inline keyboards."""
    try:
        query = update.callback_query

        # Create a new bot instance for this event loop
        local_bot = Bot(token=TELEGRAM_BOT_TOKEN)

        # Answer the callback query using the local bot
        await local_bot.answer_callback_query(callback_query_id=query.id)

        # Parse the callback data
        data = query.data.split(":", 1)  # Split on first colon only
        data_type = data[0]
        data_value = data[1] if len(data) > 1 else ""

        # Dispatch to the appropriate handler based on data_type
        if data_type == "type":
            await handle_note_type_callback(local_bot, query, data_value)
        elif data_type == "time":
            await handle_time_callback(local_bot, query, data_value)
        elif data_type == "timezone":
            await handle_timezone_callback(local_bot, query, data_value)
        else:
            logger.error(f"Unrecognized callback data type: {data_type} with value: {data_value}")
            await send_callback_response(local_bot, query, ERROR_PROCESSING_REQUEST)

    except Exception as e:
        logger.error(f"Error handling button callback: {str(e)}", exc_info=True)
        try:
            # Create a new bot instance for this event loop if not already created
            if 'local_bot' not in locals():
                local_bot = Bot(token=TELEGRAM_BOT_TOKEN)

            await send_callback_response(
                local_bot, 
                update.callback_query, 
                ERROR_PROCESSING_REQUEST
            )
        except Exception as inner_e:
            logger.error(f"Error in exception handler: {str(inner_e)}", exc_info=True)