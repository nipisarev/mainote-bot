from telegram import Update
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger
from mainote_bot.api.client import MainoteAPIClient
from mainote_bot.bot.messages import escape_markdown_v1

# Constants
UNKNOWN_APP = 'Unknown App'

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
            
            logger.info(f"Creating note for chat_id {chat_id} with category {category}")
            
            try:
                # Create the note via API
                api_client = MainoteAPIClient()
                note_response = await api_client.create_note(
                    chat_id=chat_id,
                    title=pending_note['title'],
                    content=pending_note['content'],
                    category=category,
                    status="active",
                    source="telegram"
                )
                
                note_id = note_response.get('note_id')
                logger.info(f"Successfully created note {note_id} for chat_id {chat_id}")
                
                # Clear the pending note from context
                context.user_data.pop('pending_note', None)
                
                # Send success message
                # Escape the user's content to prevent markdown parsing errors
                escaped_content = escape_markdown_v1(pending_note['content'][:200])
                truncated_suffix = '...' if len(pending_note['content']) > 200 else ''
                
                success_message = (
                    f"✅ Note saved successfully, {user_name}!\n\n"
                    f"**Category:** {category_display.get(category, category.title())}\n"
                    f"**Content:** \"{escaped_content}{truncated_suffix}\"\n\n"
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
                        text="❌ User not found. Please complete the setup using the /start command."
                    )
                elif "Server timeout" in error_message or "Unable to connect" in error_message:
                    await query.edit_message_text(
                        text="⏱️ The server is temporarily unavailable. Please try again later."
                    )
                else:
                    await query.edit_message_text(
                        text="❌ Error saving note. Please try again later.\n\n"
                    )
            
            return
        
        # Handle integration setup callbacks
        if callback_data.startswith("integration_"):
            await handle_integration_callback(update, context, callback_data)
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
                    text="❌ An error occurred while processing your request. Please try again later."
                )
        except:
            pass  # Message might be too old to edit


async def handle_integration_callback(update: Update, context: ContextTypes.DEFAULT_TYPE, callback_data: str):
    """Handle integration-related callbacks."""
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    query = update.callback_query
    api_client = MainoteAPIClient()
    
    try:
        if callback_data == "integration_setup_yes":
            # User wants to set up integrations after account creation
            await show_available_apps(update, context, api_client)
            
        elif callback_data == "integration_setup_skip":
            # User wants to skip integration setup
            await query.edit_message_text(
                text="⏭️ No problem! You can set up integrations later using the /integrations command.\n\n"
                     "Start sending me notes and I'll save them for you! 📝"
            )
            
        elif callback_data == "integration_add_new":
            # User wants to add a new integration
            await show_available_apps(update, context, api_client)
            
        elif callback_data == "integration_manage":
            # User wants to manage existing integrations
            await show_existing_integrations(update, context, api_client)
            
        elif callback_data == "integration_skip":
            # User wants to skip integration management
            await query.edit_message_text(
                text="✅ No problem! You can set up integrations anytime using the /integrations command."
            )
            
        elif callback_data.startswith("integration_select_app_"):
            # User selected an app to integrate with
            app_id = callback_data.replace("integration_select_app_", "")
            await start_app_integration(update, context, api_client, app_id)
            
        elif callback_data.startswith("integration_configure_"):
            # User is configuring an integration
            app_id = callback_data.replace("integration_configure_", "")
            await configure_integration(update, context, app_id)
            
        elif callback_data.startswith("integration_test_"):
            # User wants to test an integration
            integration_id = callback_data.replace("integration_test_", "")
            await test_integration(update, context, api_client, integration_id)
            
        elif callback_data.startswith("integration_skip_"):
            # User wants to skip a specific step
            step = callback_data.replace("integration_skip_", "")
            await handle_integration_skip(update, context, step)
            
        else:
            await query.edit_message_text(
                text="❌ Unknown action. Please try again."
            )
            
    except Exception as e:
        logger.error(f"Error handling integration callback {callback_data}: {str(e)}", exc_info=True)
        await query.edit_message_text(
            text="❌ An error occurred. Please try again later."
        )


async def show_available_apps(update: Update, context: ContextTypes.DEFAULT_TYPE, api_client: MainoteAPIClient):
    """Show available apps for integration."""
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    query = update.callback_query

    try:
        # Get available apps from the API
        apps = await api_client.get_apps()
        
        if not apps:
            await query.edit_message_text(
                text="❌ No apps available for integration at the moment. Please try again later."
            )
            return
        
        # Create message with available apps
        message = "📱 **Available Apps for Integration**\n\n"
        message += "Choose an app to integrate with:\n\n"
        
        # Create keyboard with app options
        keyboard = []
        for app in apps:
            app_name = app.get('name', UNKNOWN_APP)
            app_id = app.get('app_id')
            
            # Add app description if available
            if app.get('description'):
                message += f"**{app_name}**\n{app['description']}\n\n"
            else:
                message += f"**{app_name}**\n\n"
            
            keyboard.append([InlineKeyboardButton(
                f"🔗 {app_name}", 
                callback_data=f"integration_select_app_{app_id}"
            )])
        
        # Add skip option
        keyboard.append([InlineKeyboardButton("⏭️ Skip", callback_data="integration_skip_app_selection")])
        
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await query.edit_message_text(
            text=message,
            reply_markup=reply_markup,
            parse_mode='Markdown'
        )
        
    except Exception as e:
        logger.error(f"Error showing available apps: {str(e)}", exc_info=True)
        await query.edit_message_text(
            text="❌ Error loading apps. Please try again later."
        )


async def show_existing_integrations(update: Update, context: ContextTypes.DEFAULT_TYPE, api_client: MainoteAPIClient):
    """Show existing integrations for management."""
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    query = update.callback_query
    chat_id = str(query.message.chat_id)
    
    try:
        # Get existing integrations
        integrations_data = await api_client.get_integrations(chat_id)
        integrations = integrations_data.get('integrations', [])
        
        if not integrations:
            await query.edit_message_text(
                text="📱 You don't have any integrations yet.\n\n"
                     "Would you like to add one?",
                reply_markup=InlineKeyboardMarkup([
                    [InlineKeyboardButton("✅ Add Integration", callback_data="integration_add_new")],
                    [InlineKeyboardButton("❌ Cancel", callback_data="integration_skip")]
                ])
            )
            return
        
        # Create message with existing integrations
        message = "⚙️ **Your Integrations**\n\n"
        
        keyboard = []
        for integration in integrations:
            app_name = integration.get('app', {}).get('name', UNKNOWN_APP)
            status = integration.get('status', 'unknown')
            integration_id = integration.get('integration_id')
            
            status_emoji = {
                'active': '✅',
                'inactive': '⏸️',
                'error': '❌',
                'pending_auth': '⏳'
            }.get(status, '❓')
            
            message += f"{status_emoji} **{app_name}** - {status.title()}\n"
            
            keyboard.append([InlineKeyboardButton(
                f"{status_emoji} {app_name}", 
                callback_data=f"integration_manage_{integration_id}"
            )])
        
        # Add options
        keyboard.append([InlineKeyboardButton("➕ Add New", callback_data="integration_add_new")])
        keyboard.append([InlineKeyboardButton("❌ Close", callback_data="integration_skip")])
        
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await query.edit_message_text(
            text=message,
            reply_markup=reply_markup,
            parse_mode='Markdown'
        )
        
    except Exception as e:
        logger.error(f"Error showing existing integrations: {str(e)}", exc_info=True)
        await query.edit_message_text(
            text="❌ Error loading integrations. Please try again later."
        )


async def start_app_integration(update: Update, context: ContextTypes.DEFAULT_TYPE, api_client: MainoteAPIClient, app_id: str):
    """Start integration setup for a specific app."""
    query = update.callback_query

    try:
        # Get app details
        apps = await api_client.get_apps()
        app = next((a for a in apps if a.get('app_id') == app_id), None)
        
        if not app:
            await query.edit_message_text(
                text="❌ App not found. Please try again."
            )
            return
        
        app_name = app.get('name', UNKNOWN_APP)
        
        # Store app info in context for later use
        context.user_data['integration_setup'] = {
            'app_id': app_id,
            'app_name': app_name,
            'app': app
        }
        
        # Show app-specific setup instructions
        await show_integration_instructions(update, context, app)
        
    except Exception as e:
        logger.error(f"Error starting app integration: {str(e)}", exc_info=True)
        await query.edit_message_text(
            text="❌ Error starting integration. Please try again later."
        )


async def show_integration_instructions(update: Update, context: ContextTypes.DEFAULT_TYPE, app: dict):
    """Show integration setup instructions for a specific app."""
    from telegram import InlineKeyboardButton, InlineKeyboardMarkup
    
    query = update.callback_query
    app_name = app.get('name', UNKNOWN_APP)
    app_description = app.get('description', '')
    
    # Create instructions message
    message = f"🔧 **{app_name} Integration Setup**\n\n"
    
    if app_description:
        message += f"{app_description}\n\n"
    
    # Add general instructions based on app type
    if app_name.lower() == 'notion':
        message += (
            "**Setup Instructions:**\n\n"
            "1. **Create a Notion Integration:**\n"
            "   • Go to https://www.notion.so/my-integrations\n"
            "   • Click 'New integration'\n"
            "   • Give it a name (e.g., 'Mainote Bot')\n"
            "   • Copy the Internal Integration Token\n\n"
            "2. **Create a Database:**\n"
            "   • Create a new database in your Notion workspace\n"
            "   • Add these properties:\n"
            "     - Title (title)\n"
            "     - Content (text)\n"
            "     - Status (select: Active, Completed, Archived)\n"
            "     - Category (select: Idea, Task, Personal)\n"
            "     - Created (date)\n\n"
            "3. **Share Database with Integration:**\n"
            "   • Click 'Share' on your database\n"
            "   • Add your integration by name\n"
            "   • Copy the Database ID from the URL\n\n"
            "Ready to configure?"
        )
    else:
        message += (
            "**Setup Instructions:**\n\n"
            "Please follow the setup instructions for this app.\n"
            "You'll need to provide authentication credentials.\n\n"
            "Ready to configure?"
        )
    
    # Create keyboard
    keyboard = [
        [InlineKeyboardButton("✅ I'm Ready", callback_data=f"integration_configure_{app['app_id']}")],
        [InlineKeyboardButton("⏭️ Skip", callback_data="integration_skip_configuration")]
    ]
    
    reply_markup = InlineKeyboardMarkup(keyboard)
    
    await query.edit_message_text(
        text=message,
        reply_markup=reply_markup,
        parse_mode='Markdown'
    )


async def configure_integration(update: Update, context: ContextTypes.DEFAULT_TYPE, app_id: str):
    """Start the configuration process for an integration."""
    query = update.callback_query

    # Get app info from context
    integration_setup = context.user_data.get('integration_setup', {})
    app_name = integration_setup.get('app_name', UNKNOWN_APP)
    
    # Set up conversation state for collecting credentials
    context.user_data['integration_config_state'] = 'waiting_for_credentials'
    context.user_data['integration_config_app_id'] = app_id
    
    if app_name.lower() == 'notion':
        await query.edit_message_text(
            text="🔑 **Notion Integration Configuration**\n\n"
                 "Please send me your Notion Integration Token.\n\n"
                 "You can find it at: https://www.notion.so/my-integrations\n\n"
                 "Send the token as a message, or type 'skip' to cancel."
        )
    else:
        await query.edit_message_text(
            text=f"🔑 **{app_name} Integration Configuration**\n\n"
                 "Please send me your authentication credentials.\n\n"
                 "Send the credentials as a message, or type 'skip' to cancel."
        )


async def test_integration(update: Update, context: ContextTypes.DEFAULT_TYPE, api_client: MainoteAPIClient, integration_id: str):
    """Test an integration by creating a sample note."""
    query = update.callback_query
    chat_id = str(query.message.chat_id)
    
    try:
        # Show testing message
        await query.edit_message_text(
            text="🧪 Testing your integration...\n\n"
                 "Creating a test note to verify everything is working correctly."
        )
        
        # Test the integration
        test_result = await api_client.test_integration(chat_id, integration_id)
        
        if test_result.get('success'):
            await query.edit_message_text(
                text="✅ **Integration Test Successful!**\n\n"
                     "Your integration is working correctly. A test note has been created.\n\n"
                     "You can now start sending notes and they'll be synced automatically!"
            )
        else:
            error_message = test_result.get('message', 'Unknown error')
            await query.edit_message_text(
                text=f"❌ **Integration Test Failed**\n\n"
                     f"Error: {error_message}\n\n"
                     "Please check your configuration and try again."
            )
            
    except Exception as e:
        logger.error(f"Error testing integration: {str(e)}", exc_info=True)
        await query.edit_message_text(
            text="❌ Error testing integration. Please try again later."
        )


async def handle_integration_skip(update: Update, context: ContextTypes.DEFAULT_TYPE, step: str):
    """Handle skipping various integration steps."""
    query = update.callback_query
    
    skip_messages = {
        'app_selection': "⏭️ App selection skipped. You can set up integrations later with /integrations.",
        'configuration': "⏭️ Configuration skipped. You can complete the setup later with /integrations.",
        'testing': "⏭️ Testing skipped. Your integration has been saved but not tested."
    }
    
    message = skip_messages.get(step, "⏭️ Step skipped. You can continue later with /integrations.")
    
    await query.edit_message_text(text=message)
    
    # Clear integration setup state
    context.user_data.pop('integration_setup', None)
    context.user_data.pop('integration_config_state', None)
    context.user_data.pop('integration_config_app_id', None)