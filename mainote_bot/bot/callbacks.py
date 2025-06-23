from telegram import Update
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger


async def button_callback(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle button callbacks."""
    try:
        query = update.callback_query
        await query.answer()
        
        # For now, just acknowledge the button press
        await query.edit_message_text(
            text="✅ Thanks for your feedback! Feature coming soon."
        )
        
        logger.info(f"Button callback from user {query.from_user.id}: {query.data}")
        
    except Exception as e:
        logger.error(f"Error handling callback: {str(e)}", exc_info=True)