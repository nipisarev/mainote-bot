from telegram import Update
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger
from mainote_bot.api.client import MainoteAPIClient


async def button_callback(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle button callbacks."""
    try:
        query = update.callback_query
        await query.answer()
        
        callback_data = query.data
        chat_id = str(query.message.chat_id)
        user_name = query.from_user.first_name or "User"
        
        logger.info(f"Button callback from user {chat_id}: {callback_data}")
        
        # Handle note category selection
        if callback_data.startswith("note_category_"):
            category = callback_data.replace("note_category_", "")
            
            # Get the pending note from user context
            pending_note = context.user_data.get('pending_note')
            if not pending_note:
                await query.edit_message_text(
                    text="❌ Session expired. Please send your message again."
                )
                return
            
            # Map categories to display names
            category_display = {
                "idea": "💡 Idea",
                "task": "✅ Task", 
                "personal": "📝 Personal"
            }
            
            try:
                # Create API client and save the note
                api_client = MainoteAPIClient()
                
                note_data = await api_client.create_note(
                    chat_id=chat_id,
                    content=pending_note['content'],
                    title=pending_note['title'],
                    category=category,
                    status="active",
                    source="telegram"
                )
                
                # Clear the pending note from context
                context.user_data.pop('pending_note', None)
                
                # Send success message
                success_message = (
                    f"✅ Note saved successfully, {user_name}!\n\n"
                    f"**Category:** {category_display.get(category, category.title())}\n"
                    f"**Content:** \"{pending_note['content'][:200]}{'...' if len(pending_note['content']) > 200 else ''}\"\n\n"
                    f"📝 Your note has been saved and can be accessed in your Notion database."
                )
                
                await query.edit_message_text(
                    text=success_message,
                    parse_mode='Markdown'
                )
                
                logger.info(f"Successfully saved note for user {chat_id} with category {category}")
                
            except Exception as e:
                error_message = str(e)
                logger.error(f"Error saving note for user {chat_id}: {error_message}", exc_info=True)
                
                # Provide user-friendly error messages
                if "User not found" in error_message:
                    await query.edit_message_text(
                        text="❌ Пользователь не найден. Пожалуйста, выполните настройку с помощью команды /start."
                    )
                elif "Server timeout" in error_message or "Unable to connect" in error_message:
                    await query.edit_message_text(
                        text="⏱️ Сервер временно недоступен. Попробуйте позже."
                    )
                else:
                    await query.edit_message_text(
                        text="❌ Произошла ошибка при сохранении заметки. Попробуйте позже."
                    )
            
            return
        
        # Handle other callbacks (existing functionality)
        await query.edit_message_text(
            text="✅ Thanks for your feedback! Feature coming soon."
        )
        
    except Exception as e:
        logger.error(f"Error handling callback: {str(e)}", exc_info=True)
        try:
            if update.callback_query:
                await update.callback_query.edit_message_text(
                    text="❌ Произошла ошибка. Попробуйте позже."
                )
        except:
            pass  # Message might be too old to edit