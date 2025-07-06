from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger
from mainote_bot.bot.commands import handle_setup_flow
from mainote_bot.api.client import MainoteAPIClient


async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle incoming text messages."""
    try:
        # First check if this is part of the setup flow
        if await handle_setup_flow(update, context):
            return  # Message was handled by setup flow
        
        chat_id = str(update.effective_chat.id)
        message_text = update.message.text
        user_name = update.effective_user.first_name or "User"
        
        logger.info(f"Received message from user {chat_id}: {message_text[:100]}...")
        
        # Store the message in context for later use
        context.user_data['pending_note'] = {
            'content': message_text,
            'title': message_text,  # Use message as both content and title for now
            'chat_id': chat_id
        }
        
        # Create category selection keyboard
        keyboard = [
            [
                InlineKeyboardButton("💡 Idea", callback_data="note_category_idea"),
                InlineKeyboardButton("✅ Task", callback_data="note_category_task"),
                InlineKeyboardButton("📝 Personal", callback_data="note_category_personal")
            ]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        # Send category selection message
        await context.bot.send_message(
            chat_id=chat_id,
            text=f"📝 Got your note, {user_name}!\n\n"
                 f"**Your message:** \"{message_text[:200]}{'...' if len(message_text) > 200 else ''}\"\n\n"
                 f"Please choose a category for your note:",
            reply_markup=reply_markup,
            parse_mode='Markdown'
        )
        
        logger.info(f"Showed category selection to user {chat_id}")
        
    except Exception as e:
        logger.error(f"Error handling message: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Произошла ошибка. Попробуйте позже."
        )