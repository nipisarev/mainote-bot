from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger
from mainote_bot.notion.tasks import create_note

async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle incoming messages."""
    try:
        message = update.message
        chat_id = message.chat_id
        text = message.text

        logger.info(f"Received message from {chat_id}: {text}")

        # Save to Notion
        page = create_note(text)
        logger.info(f"Saved to Notion with ID: {page['id']}")

        # Create inline keyboard with buttons
        keyboard = [
            [
                InlineKeyboardButton("💡 Идея", callback_data=f"type:idea:{page['id']}"),
                InlineKeyboardButton("✅ Задача", callback_data=f"type:task:{page['id']}"),
                InlineKeyboardButton("📝 Личное", callback_data=f"type:personal:{page['id']}")
            ]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)

        # Send confirmation with buttons
        await context.bot.send_message(
            chat_id=chat_id,
            text="Заметка добавлена в Notion 📘\nХотите её отметить как:",
            reply_markup=reply_markup
        )

    except Exception as e:
        logger.error(f"Error handling message: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=chat_id,
            text="Произошла ошибка при сохранении заметки. Попробуйте позже."
        )