from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger
from mainote_bot.bot.commands import handle_setup_flow
from mainote_bot.api.client import MainoteAPIClient
import re


def escape_markdown_v2(text: str) -> str:
    """
    Escape special characters for Telegram's MarkdownV2 format.
    """
    # Characters that need to be escaped in MarkdownV2
    escape_chars = r'_*[]()~`>#+-=|{}.!'
    
    # Escape each character
    for char in escape_chars:
        text = text.replace(char, f'\\{char}')
    
    return text


def escape_markdown_v1(text: str) -> str:
    """
    Escape special characters for Telegram's Markdown format.
    """
    # Characters that need to be escaped in Markdown v1
    escape_chars = ['*', '_', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!']
    
    for char in escape_chars:
        text = text.replace(char, f'\\{char}')
    
    return text


async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle incoming text messages."""
    try:
        # Task enrichment flow: due date text input
        from mainote_bot.bot.callbacks import handle_task_followups
        if await handle_task_followups(update, context):
            return
        # First check if this is part of the setup flow
        if await handle_setup_flow(update, context):
            return  # Message was handled by setup flow
        
        # Check if this is part of the integration configuration flow
        if await handle_integration_config_flow(update, context):
            return  # Message was handled by integration config flow
        
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
        # Escape the user's message text to prevent markdown parsing errors
        escaped_message = escape_markdown_v1(message_text[:200])
        truncated_suffix = '...' if len(message_text) > 200 else ''
        
        await context.bot.send_message(
            chat_id=chat_id,
            text=f"📝 Got your note, {user_name}!\n\n"
                 f"**Your message:** \"{escaped_message}{truncated_suffix}\"\n\n"
                 f"Please choose a category for your note:",
            reply_markup=reply_markup,
            parse_mode='Markdown'
        )
        
        logger.info(f"Showed category selection to user {chat_id}")
        
    except Exception as e:
        logger.error(f"Error handling message: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Sorry, something went wrong. Please try again later."
        )


async def handle_integration_config_flow(update: Update, context: ContextTypes.DEFAULT_TYPE) -> bool:
    """
    Handle the integration configuration conversation flow.
    Returns True if message was handled as part of integration config, False otherwise.
    """
    if 'integration_config_state' not in context.user_data:
        return False
        
    chat_id = str(update.effective_chat.id)
    message_text = update.message.text.strip()
    config_state = context.user_data.get('integration_config_state')
    app_id = context.user_data.get('integration_config_app_id')
    
    from mainote_bot.api.client import MainoteAPIClient
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    api_client = MainoteAPIClient()
    
    try:
        if config_state == 'waiting_for_credentials':
            # Check if user wants to skip
            if message_text.lower() == 'skip':
                await context.bot.send_message(
                    chat_id=chat_id,
                    text="⏭️ Configuration skipped. You can complete the setup later with /integrations."
                )
                # Clear integration config state
                context.user_data.pop('integration_config_state', None)
                context.user_data.pop('integration_config_app_id', None)
                return True
            
            # Get app info from context
            integration_setup = context.user_data.get('integration_setup', {})
            app_name = integration_setup.get('app_name', 'Unknown App')
            
            # Handle Notion integration
            if app_name.lower() == 'notion':
                # The message should be the Notion integration token
                token = message_text
                
                # Update state to collect database ID
                context.user_data['integration_config_state'] = 'waiting_for_database_id'
                context.user_data['integration_notion_token'] = token
                
                await context.bot.send_message(
                    chat_id=chat_id,
                    text="✅ Token received!\n\n"
                         "Now please send me your Notion Database ID.\n\n"
                         "You can find it in your database URL:\n"
                         "https://www.notion.so/[workspace]/[DATABASE_ID]?v=...\n\n"
                         "Send just the Database ID, or type 'skip' to cancel."
                )
                
                return True
            
            else:
                # Handle other integrations generically
                await create_generic_integration(
                    update, context, api_client, app_id, app_name, message_text
                )
                return True
                
        elif config_state == 'waiting_for_database_id':
            # Check if user wants to skip
            if message_text.lower() == 'skip':
                await context.bot.send_message(
                    chat_id=chat_id,
                    text="⏭️ Configuration skipped. You can complete the setup later with /integrations."
                )
                # Clear integration config state
                context.user_data.pop('integration_config_state', None)
                context.user_data.pop('integration_config_app_id', None)
                context.user_data.pop('integration_notion_token', None)
                return True
            
            # The message should be the database ID
            database_id = message_text
            token = context.user_data.get('integration_notion_token')
            
            # Create the Notion integration
            await create_notion_integration(
                update, context, api_client, app_id, token, database_id
            )
            return True
            
    except Exception as e:
        logger.error(f"Error in integration config flow: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=chat_id,
            text="❌ An error occurred during configuration. Please try again with /integrations."
        )
        # Clear integration config state
        context.user_data.pop('integration_config_state', None)
        context.user_data.pop('integration_config_app_id', None)
        context.user_data.pop('integration_notion_token', None)
        return True
    
    return False


async def create_notion_integration(update: Update, context: ContextTypes.DEFAULT_TYPE, api_client: MainoteAPIClient, app_id: str, token: str, database_id: str):
    """Create a Notion integration with the provided credentials."""
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    chat_id = str(update.effective_chat.id)
    
    try:
        # Show processing message
        processing_msg = await context.bot.send_message(
            chat_id=chat_id,
            text="⚙️ Creating your Notion integration..."
        )
        
        # Create the integration
        integration_data = await api_client.create_integration(
            chat_id=chat_id,
            app_id=app_id,
            auth_type="api_key",
            auth_data={
                "api_key": token,
                "database_id": database_id
            },
            config={
                "sync_frequency": "realtime",
                "auto_sync": True
            }
        )
        
        # Delete processing message
        await context.bot.delete_message(chat_id=chat_id, message_id=processing_msg.message_id)
        
        integration_id = integration_data.get('integration_id')
        
        # Create success message with test option
        success_message = (
            "✅ **Notion Integration Created!**\n\n"
            "Your Notion integration has been successfully configured.\n\n"
            "Would you like to test it now?"
        )
        
        keyboard = [
            [InlineKeyboardButton("🧪 Test Integration", callback_data=f"integration_test_{integration_id}")],
            [InlineKeyboardButton("⏭️ Skip Test", callback_data="integration_skip_testing")]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await context.bot.send_message(
            chat_id=chat_id,
            text=success_message,
            reply_markup=reply_markup,
            parse_mode='Markdown'
        )
        
        logger.info(f"Successfully created Notion integration for user {chat_id}")
        
    except Exception as e:
        # Delete processing message if it exists
        try:
            await context.bot.delete_message(chat_id=chat_id, message_id=processing_msg.message_id)
        except:
            pass
        
        error_message = str(e)
        logger.error(f"Error creating Notion integration: {error_message}")
        
        if "Integration already exists" in error_message:
            await context.bot.send_message(
                chat_id=chat_id,
                text="❌ You already have a Notion integration configured. Please use /integrations to manage it."
            )
        elif "User or app not found" in error_message:
            await context.bot.send_message(
                chat_id=chat_id,
                text="❌ Setup issue. Please ensure you've completed account setup with /start."
            )
        else:
            await context.bot.send_message(
                chat_id=chat_id,
                text=f"❌ Failed to create integration: {error_message}\n\n"
                     "Please check your credentials and try again."
            )
    
    finally:
        # Clear integration config state
        context.user_data.pop('integration_config_state', None)
        context.user_data.pop('integration_config_app_id', None)
        context.user_data.pop('integration_notion_token', None)
        context.user_data.pop('integration_setup', None)


async def create_generic_integration(update: Update, context: ContextTypes.DEFAULT_TYPE, api_client: MainoteAPIClient, app_id: str, app_name: str, credentials: str):
    """Create a generic integration with the provided credentials."""
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    chat_id = str(update.effective_chat.id)
    
    try:
        # Show processing message
        processing_msg = await context.bot.send_message(
            chat_id=chat_id,
            text=f"⚙️ Creating your {app_name} integration..."
        )
        
        # Create the integration
        integration_data = await api_client.create_integration(
            chat_id=chat_id,
            app_id=app_id,
            auth_type="api_key",
            auth_data={
                "api_key": credentials
            },
            config={
                "sync_frequency": "realtime",
                "auto_sync": True
            }
        )
        
        # Delete processing message
        await context.bot.delete_message(chat_id=chat_id, message_id=processing_msg.message_id)
        
        integration_id = integration_data.get('integration_id')
        
        # Create success message with test option
        success_message = (
            f"✅ **{app_name} Integration Created!**\n\n"
            f"Your {app_name} integration has been successfully configured.\n\n"
            "Would you like to test it now?"
        )
        
        keyboard = [
            [InlineKeyboardButton("🧪 Test Integration", callback_data=f"integration_test_{integration_id}")],
            [InlineKeyboardButton("⏭️ Skip Test", callback_data="integration_skip_testing")]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await context.bot.send_message(
            chat_id=chat_id,
            text=success_message,
            reply_markup=reply_markup,
            parse_mode='Markdown'
        )
        
        logger.info(f"Successfully created {app_name} integration for user {chat_id}")
        
    except Exception as e:
        # Delete processing message if it exists
        try:
            await context.bot.delete_message(chat_id=chat_id, message_id=processing_msg.message_id)
        except:
            pass
        
        error_message = str(e)
        logger.error(f"Error creating {app_name} integration: {error_message}")
        
        if "Integration already exists" in error_message:
            await context.bot.send_message(
                chat_id=chat_id,
                text=f"❌ You already have a {app_name} integration configured. Please use /integrations to manage it."
            )
        elif "User or app not found" in error_message:
            await context.bot.send_message(
                chat_id=chat_id,
                text="❌ Setup issue. Please ensure you've completed account setup with /start."
            )
        else:
            await context.bot.send_message(
                chat_id=chat_id,
                text=f"❌ Failed to create integration: {error_message}\n\n"
                     "Please check your credentials and try again."
            )
    
    finally:
        # Clear integration config state
        context.user_data.pop('integration_config_state', None)
        context.user_data.pop('integration_config_app_id', None)
        context.user_data.pop('integration_setup', None)