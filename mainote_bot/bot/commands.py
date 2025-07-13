from telegram import Update
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger
from mainote_bot.api.client import MainoteAPIClient
import re
import secrets
import string


async def start_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /start command with user setup flow."""
    try:
        chat_id = str(update.effective_chat.id)
        user_name = update.effective_user.first_name or "there"
        
        # Initialize API client
        api_client = MainoteAPIClient()
        
        # Check if user already exists
        try:
            existing_user = await api_client.get_user_by_chat_id(chat_id)
            
            if existing_user:
                # User already exists and is connected
                welcome_message = (
                    f"👋 Welcome back {user_name}!\n\n"
                    f"You're logged in as: {existing_user.get('email', 'Unknown')}\n\n"
                    "I'm ready to help you save your notes! Simply send me any text message and I'll save it for you.\n\n"
                    "Commands:\n"
                    "/start - Show this welcome message\n"
                    "/help - Get help and information"
                )
                
                await context.bot.send_message(
                    chat_id=chat_id,
                    text=welcome_message
                )
                
                logger.info(f"Returning user {existing_user.get('email')} welcomed back to chat {chat_id}")
                return
                
        except Exception as e:
            # Log the error but continue with setup flow
            logger.error(f"Error checking existing user for chat {chat_id}: {str(e)}")
            
            # Send error message to user and stop
            await context.bot.send_message(
                chat_id=chat_id,
                text="Sorry, I'm having trouble connecting to the server right now. Please try again in a few minutes."
            )
            return
        
        # New user - start setup flow
        welcome_message = (
            f"👋 Hello {user_name}! Welcome to Mainote Bot!\n\n"
            "I'm here to help you save and organize your notes.\n\n"
            "To get started, I just need your email address.\n"
            "Please send me your email address."
        )
        
        # Set conversation state
        context.user_data['setup_state'] = 'waiting_for_email'
        context.user_data['chat_id'] = chat_id
        
        await context.bot.send_message(
            chat_id=chat_id,
            text=welcome_message
        )
        
        logger.info(f"Started setup flow for new user in chat {chat_id}")
        
    except Exception as e:
        logger.error(f"Error in start command: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Sorry, something went wrong. Please try again later."
        )


async def help_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /help command."""
    try:
        chat_id = update.effective_chat.id
        
        help_message = (
            "ℹ️ **Mainote Bot Help**\n\n"
            "I'm a simple note-taking bot. Here's how to use me:\n\n"
            "📝 **Taking Notes:**\n"
            "• Send me any text message\n"
            "• I'll automatically save it as a note\n\n"
            "🎯 **Commands:**\n"
            "• /start - Show welcome message and set up account\n"
            "• /help - Show this help message\n"
            "• /integrations - Set up integrations with external apps\n"
            "• /reset - Reset setup process if you get stuck\n\n"
            "That's it! Keep it simple and start taking notes! 📚"
        )
        
        await context.bot.send_message(
            chat_id=chat_id,
            text=help_message,
            parse_mode="Markdown"
        )
        
        logger.info(f"Sent help message to user {chat_id}")
        
    except Exception as e:
        logger.error(f"Error in help command: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Sorry, something went wrong. Please try again later."
        )


async def integrations_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /integrations command to set up integrations."""
    try:
        chat_id = str(update.effective_chat.id)
        user_name = update.effective_user.first_name or "there"
        
        # Initialize API client
        api_client = MainoteAPIClient()
        
        # Check if user exists
        try:
            existing_user = await api_client.get_user_by_chat_id(chat_id)
            if not existing_user:
                await context.bot.send_message(
                    chat_id=chat_id,
                    text="❌ You need to complete setup first. Please use /start to create your account."
                )
                return
        except Exception as e:
            logger.error(f"Error checking user for integrations: {str(e)}")
            await context.bot.send_message(
                chat_id=chat_id,
                text="Sorry, I'm having trouble connecting to the server. Please try again later."
            )
            return
        
        # Start the integration setup flow
        await start_integration_setup(update, context, api_client)
        
    except Exception as e:
        logger.error(f"Error in integrations command: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Sorry, something went wrong. Please try again later."
        )


async def start_integration_setup(update: Update, context: ContextTypes.DEFAULT_TYPE, api_client: MainoteAPIClient):
    """Start the integration setup flow."""
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    chat_id = str(update.effective_chat.id)
    user_name = update.effective_user.first_name or "there"
    
    # Check existing integrations
    try:
        existing_integrations = await api_client.get_integrations(chat_id)
        integration_count = len(existing_integrations.get('integrations', []))
    except Exception as e:
        logger.error(f"Error getting existing integrations: {str(e)}")
        integration_count = 0
    
    # Create intro message
    intro_message = (
        f"🔗 **Integration Setup**\n\n"
        f"Hey {user_name}! Let's set up integrations to sync your notes with external apps.\n\n"
        f"**Why integrate?**\n"
        f"• Sync tasks to your favorite productivity apps\n"
        f"• Keep everything organized in one place\n"
        f"• Never lose important notes again\n\n"
    )
    
    if integration_count > 0:
        intro_message += f"You currently have {integration_count} integration(s) configured.\n\n"
    
    intro_message += "Would you like to:"
    
    # Create keyboard buttons
    keyboard = [
        [InlineKeyboardButton("📱 Add New Integration", callback_data="integration_add_new")],
    ]
    
    if integration_count > 0:
        keyboard.append([InlineKeyboardButton("⚙️ Manage Existing", callback_data="integration_manage")])
    
    keyboard.append([InlineKeyboardButton("❌ Skip", callback_data="integration_skip")])
    
    reply_markup = InlineKeyboardMarkup(keyboard)
    
    await context.bot.send_message(
        chat_id=chat_id,
        text=intro_message,
        reply_markup=reply_markup,
        parse_mode='Markdown'
    )
    
    logger.info(f"Started integration setup for user {chat_id}")


async def suggest_integrations_setup(update: Update, context: ContextTypes.DEFAULT_TYPE, api_client: MainoteAPIClient):
    """Suggest setting up integrations after successful account creation."""
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    chat_id = str(update.effective_chat.id)
    user_name = update.effective_user.first_name or "there"
    
    # Create suggestion message
    suggestion_message = (
        f"🔗 **Want to sync your notes to external apps?**\n\n"
        f"Set up integrations to automatically sync your notes with apps like Notion, TickTick, or Jira.\n\n"
        f"**Benefits:**\n"
        f"• Keep all your notes in one place\n"
        f"• Never lose important information\n"
        f"• Work with your favorite productivity tools\n\n"
        f"Would you like to set up integrations now?"
    )
    
    # Create keyboard buttons
    keyboard = [
        [InlineKeyboardButton("✅ Yes, Set Up Integrations", callback_data="integration_setup_yes")],
        [InlineKeyboardButton("⏭️ Skip for Now", callback_data="integration_setup_skip")]
    ]
    
    reply_markup = InlineKeyboardMarkup(keyboard)
    
    await context.bot.send_message(
        chat_id=chat_id,
        text=suggestion_message,
        reply_markup=reply_markup,
        parse_mode='Markdown'
    )
    
    logger.info(f"Suggested integration setup for user {chat_id}")


def is_valid_email(email: str) -> bool:
    """Validate email format."""
    pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    return bool(re.match(pattern, email))


def generate_secure_password(length: int = 12) -> str:
    """Generate a secure random password."""
    # Use a mix of letters, digits, and some safe special characters
    alphabet = string.ascii_letters + string.digits + "!@#$%^&*"
    password = ''.join(secrets.choice(alphabet) for _ in range(length))
    return password


async def handle_setup_flow(update: Update, context: ContextTypes.DEFAULT_TYPE) -> bool:
    """
    Handle the user setup conversation flow.
    Returns True if message was handled as part of setup, False otherwise.
    """
    if 'setup_state' not in context.user_data:
        return False
        
    chat_id = str(update.effective_chat.id)
    message_text = update.message.text.strip()
    setup_state = context.user_data.get('setup_state')
    
    api_client = MainoteAPIClient()
    
    try:
        if setup_state == 'waiting_for_email':
            # Validate email format
            if not is_valid_email(message_text):
                await context.bot.send_message(
                    chat_id=chat_id,
                    text="❌ That doesn't look like a valid email address.\nPlease send me a valid email address (e.g., user@example.com):"
                )
                return True
            
            email = message_text
            
            # Show processing message
            processing_msg = await context.bot.send_message(
                chat_id=chat_id,
                text="� Setting up your account..."
            )
            
            try:
                # Generate a secure password automatically
                auto_password = generate_secure_password()
                
                # Since we've already checked that no user exists for this chat_id,
                # we know this is a new user. Create user with auto-generated password.
                user_data = await api_client.create_user(email, auto_password)
                logger.info(f"Created new user {email} with auto-generated password")
                
                # Now link the user to chat_id via auth (this just sets the chat_id)
                auth_result = await api_client.authenticate_user(chat_id, email, auto_password)
                
                logger.info(f"Auth result: {auth_result}")  # Debug log
                
                if not auth_result or 'user_id' not in auth_result:
                    logger.error(f"Authentication failed: auth_result={auth_result}")
                    raise Exception("Authentication failed: unable to get user_id")
                
                user_id = auth_result['user_id']
                logger.info(f"Authenticated user {email} with user_id {user_id}")
                
                # Create user settings to associate user with chat_id
                try:
                    await api_client.create_user_settings(user_id, chat_id)
                    logger.info(f"Created user settings for user_id {user_id}, chat_id {chat_id}")
                except Exception as settings_error:
                    # Log error but don't fail the whole setup - user is already created
                    logger.error(f"Failed to create user settings for user_id {user_id}, chat_id {chat_id}: {str(settings_error)}")
                
                # Clear setup state
                context.user_data.clear()
                
                # Delete processing message
                await context.bot.delete_message(chat_id=chat_id, message_id=processing_msg.message_id)
                
                # Send success message
                success_message = (
                    "✅ Great! You're all set up!\n\n"
                    f"📧 Email: {email}\n"
                    f"💬 Chat ID: {chat_id}\n\n"
                    "Your account has been created automatically!\n\n"
                    "You can now start sending me notes! Just type any message and I'll save it for you.\n\n"
                    "Commands:\n"
                    "/start - Show welcome message\n"
                    "/help - Get help and information\n"
                    "/integrations - Set up integrations with external apps"
                )
                
                await context.bot.send_message(
                    chat_id=chat_id,
                    text=success_message
                )
                
                # Suggest setting up integrations
                await suggest_integrations_setup(update, context, api_client)
                
                logger.info(f"Successfully set up user {email} for chat {chat_id} with auto-generated password")
                return True
                
            except Exception as e:
                # Delete processing message
                try:
                    await context.bot.delete_message(chat_id=chat_id, message_id=processing_msg.message_id)
                except:
                    pass  # Ignore if we can't delete
                
                error_message = str(e)
                
                if "already connected" in error_message:
                    await context.bot.send_message(
                        chat_id=chat_id,
                        text="❌ This chat is already connected to another user account.\n\nIf you believe this is an error, please contact support."
                    )
                    # Clear setup state
                    context.user_data.clear()
                    
                elif "Server timeout" in error_message or "Unable to connect" in error_message:
                    await context.bot.send_message(
                        chat_id=chat_id,
                        text="❌ Server connection problem. Please try again in a few minutes.\n\nSend /start to try again."
                    )
                    # Clear setup state
                    context.user_data.clear()
                    
                else:
                    await context.bot.send_message(
                        chat_id=chat_id,
                        text=f"❌ Setup failed: {error_message}\n\nSend /start to try again."
                    )
                    # Clear setup state
                    context.user_data.clear()
                
                logger.error(f"Setup failed for chat {chat_id}: {error_message}")
                return True
                
    except Exception as e:
        logger.error(f"Unexpected error in setup flow for chat {chat_id}: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=chat_id,
            text="❌ An unexpected error occurred. Please send /start to try again."
        )
        # Clear setup state
        context.user_data.clear()
        return True
    
    return False


async def reset_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /reset command to clear setup state."""
    try:
        chat_id = str(update.effective_chat.id)
        
        # Clear any setup state
        context.user_data.clear()
        
        reset_message = (
            "🔄 Your setup has been reset.\n\n"
            "Send /start to begin the setup process again."
        )
        
        await context.bot.send_message(
            chat_id=chat_id,
            text=reset_message
        )
        
        logger.info(f"Reset setup state for user in chat {chat_id}")
        
    except Exception as e:
        logger.error(f"Error in reset command: {str(e)}", exc_info=True)
        await context.bot.send_message(
            chat_id=update.effective_chat.id,
            text="Sorry, something went wrong. Please try again later."
        )
