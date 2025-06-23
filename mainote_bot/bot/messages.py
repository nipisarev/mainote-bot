from telegram import Update
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger
from mainote_bot.bot.commands import handle_setup_flow


async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle incoming text messages."""
    try:
        # First check if this is part of the setup flow
        if await handle_setup_flow(update, context):
            return  # Message was handled by setup flow
        
        chat_id = update.effective_chat.id
        message_text = update.message.text
        user_name = update.effective_user.first_name or "User"
        
        # Check if user is set up
        # For now, if no setup_state exists, assume user is set up
        # In a real implementation, you might want to verify with the API
        
        # For now, just acknowledge the message
        response_message = (
            f"✅ Got your note, {user_name}!\n\n"
            f"Your message: \"{message_text}\"\n\n"
            "Note saved successfully! 📝"
        )
        
        await context.bot.send_message(
            chat_id=chat_id,
            text=response_message
        )
        
        logger.info(f"Processed message from user {chat_id}: {message_text[:50]}...")
        
    except Exception as e:
        logger.error(f"Error handling message: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Sorry, I couldn't save your note. Please try again later."
        )