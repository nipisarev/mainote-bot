from telegram import Update
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger


async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle incoming text messages."""
    try:
        message = update.message
        chat_id = message.chat_id
        text = message.text

        logger.info(f"Received message from {chat_id}: {text}")

        # For now, just acknowledge the message
        # TODO: Implement note saving functionality via API
        await context.bot.send_message(
            chat_id=chat_id,
            text=f"Received your note: \"{text}\"\n\n"
                 f"Note saving functionality is currently under development. "
                 f"Use /help command for assistance."
        )

    except Exception as e:
        logger.error(f"Error handling message: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="An error occurred while processing the message. Please try again later."
        )
