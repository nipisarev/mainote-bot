from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import ContextTypes
from mainote_bot.utils.logging import logger
from mainote_bot.api.client import MainoteAPIClient
from mainote_bot.bot.messages import escape_markdown_v1

# Constants
UNKNOWN_APP = 'Unknown App'
NOTES_PER_PAGE = 10

def create_note_action_buttons(note_id: str, note_num: int, total_notes: int) -> InlineKeyboardMarkup:
    """Create action buttons for a specific note."""
    buttons = []
    
    # First row: Done, Delete
    buttons.append([
        InlineKeyboardButton("✅ Done", callback_data=f"note_action_done_{note_id}_{note_num}"),
        InlineKeyboardButton("🗑️ Delete", callback_data=f"note_action_delete_{note_id}_{note_num}")
    ])
    
    # Second row: Close, Next (if not last note)
    second_row = [InlineKeyboardButton("❌ Close", callback_data="notes_close")]
    if note_num < total_notes:
        second_row.append(InlineKeyboardButton("➡️ Next", callback_data=f"note_action_next_{note_id}_{note_num}"))
    
    buttons.append(second_row)
    
    return InlineKeyboardMarkup(buttons)

def create_notes_buttons_from_storage(notes_map: dict, page: int = 0) -> InlineKeyboardMarkup:
    """Create inline keyboard buttons for note browsing with pagination."""
    if not notes_map:
        return InlineKeyboardMarkup([])
    
    buttons = []
    note_numbers = sorted(notes_map.keys())
    
    # Calculate pagination
    start_idx = page * NOTES_PER_PAGE
    end_idx = min(start_idx + NOTES_PER_PAGE, len(note_numbers))
    
    # Create number buttons in rows of 5
    current_row = []
    for i in range(start_idx, end_idx):
        note_num = note_numbers[i]
        current_row.append(InlineKeyboardButton(
            str(note_num), 
            callback_data=f"note_browse_{note_num}_{notes_map[note_num]}"
        ))
        
        # Add row every 5 buttons
        if len(current_row) == 5:
            buttons.append(current_row)
            current_row = []
    
    # Add remaining buttons
    if current_row:
        buttons.append(current_row)
    
    # Add navigation buttons
    nav_buttons = []
    total_pages = (len(note_numbers) + NOTES_PER_PAGE - 1) // NOTES_PER_PAGE
    
    if page > 0:
        nav_buttons.append(InlineKeyboardButton("⬅️ Previous", callback_data=f"notes_page_{page-1}"))
    
    if page < total_pages - 1:
        nav_buttons.append(InlineKeyboardButton("➡️ Next", callback_data=f"notes_page_{page+1}"))
    
    if nav_buttons:
        buttons.append(nav_buttons)
    
    # Add close button
    buttons.append([InlineKeyboardButton("❌ Close", callback_data="notes_close")])
    
    return InlineKeyboardMarkup(buttons)

async def button_callback(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle button callbacks."""
    try:
        query = update.callback_query
        await query.answer()
        
        callback_data = query.data
        chat_id = str(query.message.chat_id)
        user_name = query.from_user.first_name or "User"
        
        logger.info(f"Button callback from user {chat_id}: {callback_data}")
        
        # Handle note browsing callbacks
        if callback_data.startswith("note_browse_"):
            await handle_note_browse(update, context, callback_data)
            return
        
        # Handle note action callbacks
        if callback_data.startswith("note_action_"):
            await handle_note_action(update, context, callback_data)
            return
        
        # Handle notes pagination
        if callback_data.startswith("notes_page_"):
            await handle_notes_page(update, context, callback_data)
            return
        
        # Handle notes close
        if callback_data == "notes_close":
            await query.edit_message_text(
                text="👋 Note browsing closed. Have a great day!"
            )
            return

        # Task follow-up: due date set/skip
        if callback_data in ("task_due_set", "task_due_skip"):
            note_id = context.user_data.get('last_created_note_id')
            if not note_id:
                await query.edit_message_text(text="Session expired. Send a new task.")
                return
            if callback_data == "task_due_set":
                await query.edit_message_text(text="📅 Send due date in format dd.mm.yyyy hh:mm (UTC)")
            else:
                # Skip due date, go to effort
                await _prompt_effort_buttons(query, context, user_name)
            return

        # Task follow-up: effort
        if callback_data.startswith("task_effort_"):
            note_id = context.user_data.get('last_created_note_id')
            if not note_id:
                await query.edit_message_text(text="Session expired. Send a new task.")
                return
            minutes = int(callback_data.split("_")[-1])
            try:
                api = MainoteAPIClient()
                await api.update_note(note_id, chat_id, metadata={"effort_min": minutes})
                await _prompt_priority_buttons(query, context, user_name)
            except Exception as e:
                await query.edit_message_text(text=f"❌ Failed to set effort: {e}")
            return

        # Task follow-up: priority
        if callback_data.startswith("task_priority_"):
            note_id = context.user_data.get('last_created_note_id')
            if not note_id:
                await query.edit_message_text(text="Session expired. Send a new task.")
                return
            value = int(callback_data.split("_")[-1])
            try:
                api = MainoteAPIClient()
                await api.update_note(note_id, chat_id, metadata={"priority": value})
                await _finish_task_enrichment(query, context)
            except Exception as e:
                await query.edit_message_text(text=f"❌ Failed to set priority: {e}")
            return
        
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
                
                # If it's a task, start follow-up flow for due date / effort / priority
                if category == "task" and note_id:
                    context.user_data['last_created_note_id'] = note_id
                    # Ask about due date
                    keyboard = [
                        [
                            InlineKeyboardButton("📅 Set date", callback_data="task_due_set"),
                            InlineKeyboardButton("❌ No", callback_data="task_due_skip")
                        ]
                    ]
                    escaped_content = escape_markdown_v1(pending_note['content'][:200])
                    truncated_suffix = '...' if len(pending_note['content']) > 200 else ''
                    await query.edit_message_text(
                        text=(
                            f"✅ Task saved, {user_name}!\n\n"
                            f"**Content:** \"{escaped_content}{truncated_suffix}\"\n\n"
                            f"Do you want to set a due date?"
                        ),
                        reply_markup=InlineKeyboardMarkup(keyboard),
                        parse_mode='Markdown'
                    )
                else:
                    # Send success message for non-task
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

async def _prompt_effort_buttons(query, context, user_name: str):
    buttons = [
        [
            InlineKeyboardButton("XS • 15m", callback_data="task_effort_15"),
            InlineKeyboardButton("S • 30m", callback_data="task_effort_30"),
            InlineKeyboardButton("M • 2h", callback_data="task_effort_120"),
        ],
        [
            InlineKeyboardButton("L • 8h", callback_data="task_effort_480"),
            InlineKeyboardButton("XL • 2d", callback_data="task_effort_2880"),
        ]
    ]
    await query.edit_message_text(
        text=f"⏱️ Select estimated effort, {user_name}:",
        reply_markup=InlineKeyboardMarkup(buttons)
    )

async def _prompt_priority_buttons(query, context, user_name: str):
    buttons = [
        [
            InlineKeyboardButton("⬇️ Lowest", callback_data="task_priority_10"),
            InlineKeyboardButton("↓ Low", callback_data="task_priority_20"),
        ],
        [
            InlineKeyboardButton("• Normal", callback_data="task_priority_0"),
            InlineKeyboardButton("↑ High", callback_data="task_priority_30"),
            InlineKeyboardButton("⬆️ Highest", callback_data="task_priority_40"),
        ]
    ]
    await query.edit_message_text(
        text=f"🎯 Set priority?",
        reply_markup=InlineKeyboardMarkup(buttons)
    )

async def _finish_task_enrichment(query, context):
    context.user_data.pop('last_created_note_id', None)
    await query.edit_message_text(text="✅ Task details updated.")

async def _parse_ddmmyyyy_hhmm(s: str) -> str | None:
    import re, datetime
    m = re.match(r"^(\d{2})\.(\d{2})\.(\d{4})\s+(\d{2}):(\d{2})$", s)
    if not m:
        return None
    dd, mm, yyyy, hh, mi = map(int, m.groups())
    try:
        dt = datetime.datetime(yyyy, mm, dd, hh, mi, tzinfo=datetime.timezone.utc)
        return dt.isoformat()
    except Exception:
        return None

async def handle_task_followups(update: Update, context: ContextTypes.DEFAULT_TYPE) -> bool:
    """Entry from messages.handle_message to capture date input when needed."""
    if 'last_created_note_id' not in context.user_data:
        return False
    if not update.message or not update.message.text:
        return False

    chat_id = str(update.effective_chat.id)
    text = update.message.text.strip()
    note_id = context.user_data.get('last_created_note_id')
    api = MainoteAPIClient()

    # Expecting dd.mm.yyyy hh:mm format
    iso = await _parse_ddmmyyyy_hhmm(text)
    if not iso:
        await context.bot.send_message(chat_id=chat_id, text="❌ Invalid format. Use dd.mm.yyyy hh:mm")
        return True

    try:
        await api.update_note(note_id, chat_id, metadata={"due_at": iso})
        # After due date set, prompt effort
        await context.bot.send_message(chat_id=chat_id, text="📅 Due date set.")
        # Prompt effort via a fresh message with buttons
        keyboard = [
            [
                InlineKeyboardButton("XS • 15m", callback_data="task_effort_15"),
                InlineKeyboardButton("S • 30m", callback_data="task_effort_30"),
                InlineKeyboardButton("M • 2h", callback_data="task_effort_120"),
            ],
            [
                InlineKeyboardButton("L • 8h", callback_data="task_effort_480"),
                InlineKeyboardButton("XL • 2d", callback_data="task_effort_2880"),
            ]
        ]
        await context.bot.send_message(chat_id=chat_id, text="⏱️ Select estimated effort:", reply_markup=InlineKeyboardMarkup(keyboard))
    except Exception as e:
        await context.bot.send_message(chat_id=chat_id, text=f"❌ Failed to set due date: {e}")
    return True


async def handle_note_browse(update: Update, context: ContextTypes.DEFAULT_TYPE, callback_data: str):
    """Handle note browsing callback."""
    try:
        query = update.callback_query
        chat_id = str(query.message.chat_id)
        
        # Parse callback data: note_browse_{note_num}_{note_id}
        parts = callback_data.split("_", 3)
        if len(parts) < 4:
            logger.error(f"Invalid note browse callback data: {callback_data}")
            return
        
        note_num = int(parts[2])
        note_id = parts[3]
        
        logger.info(f"User {chat_id} browsing note {note_num} (ID: {note_id})")
        
        # Get note details from API
        api_client = MainoteAPIClient()
        note_data = await api_client.get_note_by_id(note_id, chat_id)
        
        # Get notes map from bot storage to determine total notes
        bot = context.bot
        notes_map = getattr(bot, '_notes_storage', {}).get(chat_id, {})
        total_notes = len(notes_map)
        
        # Format note display
        title = note_data.get('title', 'Untitled')
        content = note_data.get('content', '')
        category = note_data.get('category', 'general')
        status = note_data.get('status', 'active')
        
        # Truncate content if too long
        if len(content) > 500:
            content = content[:500] + "..."
        
        category_emojis = {
            'task': '✅',
            'idea': '💡',
            'personal': '🏖',
            'work': '💼',
            'general': '📄'
        }
        
        emoji = category_emojis.get(category, '📄')
        
        message_text = (
            f"📝 **Note {note_num}**\n\n"
            f"{emoji} **{title}**\n"
            f"Category: {category.title()}\n"
            f"Status: {status.title()}\n\n"
            f"**Content:**\n{content}"
        )
        # Append extras if present
        extras = []
        due_at = note_data.get('due_at')
        if due_at:
            extras.append(f"⏰ {str(due_at)[:10]}")
        effort_min = note_data.get('effort_min')
        if isinstance(effort_min, int) and effort_min > 0:
            extras.append(f"⏳ {effort_min}m")
        priority = note_data.get('priority')
        if isinstance(priority, int) and priority != 0:
            extras.append(f"⭐ {priority}")

        if extras:
            message_text = message_text + "\n\n" + " • ".join(extras)
        
        # Create action buttons
        keyboard = create_note_action_buttons(note_id, note_num, total_notes)
        
        await query.edit_message_text(
            text=message_text,
            reply_markup=keyboard,
            parse_mode='Markdown'
        )
        
    except Exception as e:
        logger.error(f"Error in handle_note_browse: {str(e)}", exc_info=True)
        await query.edit_message_text(
            text="❌ Error loading note. Please try again later."
        )

async def handle_note_action(update: Update, context: ContextTypes.DEFAULT_TYPE, callback_data: str):
    """Handle note action callbacks (Done, Delete, Next)."""
    try:
        query = update.callback_query
        chat_id = str(query.message.chat_id)
        
        # Parse callback data: note_action_{action}_{note_id}_{note_num}
        parts = callback_data.split("_", 4)
        if len(parts) != 5:
            logger.error(f"Invalid note action callback data: {callback_data}")
            return
        
        action = parts[2]
        note_id = parts[3]
        note_num = int(parts[4])
        
        logger.info(f"User {chat_id} performing action {action} on note {note_id}")
        
        api_client = MainoteAPIClient()
        
        if action == "done":
            # Mark note as completed
            await api_client.update_note_status(note_id, "completed", chat_id)
            await query.edit_message_text(
                text=f"✅ Note {note_num} marked as completed!\n\n"
                     f"Great job! The note has been marked as done."
            )
            
        elif action == "delete":
            # Delete the note
            await api_client.delete_note(note_id, chat_id)
            await query.edit_message_text(
                text=f"🗑️ Note {note_num} deleted successfully!\n\n"
                     f"The note has been permanently removed."
            )
            
        elif action == "next":
            # Show next note
            bot = context.bot
            notes_map = getattr(bot, '_notes_storage', {}).get(chat_id, {})
            note_numbers = sorted(notes_map.keys())
            
            logger.info(f"Looking for next note after {note_num}. Available notes: {note_numbers}")
            
            # Find next note
            try:
                current_index = note_numbers.index(note_num)
                if current_index < len(note_numbers) - 1:
                    next_note_num = note_numbers[current_index + 1]
                    next_note_id = notes_map[next_note_num]
                    
                    logger.info(f"Found next note: {next_note_num} -> {next_note_id}")
                    
                    # Simulate browsing the next note
                    await handle_note_browse(update, context, f"note_browse_{next_note_num}_{next_note_id}")
                    return
                else:
                    await query.edit_message_text(
                        text="📝 You've reached the last note!\n\n"
                             "Great job browsing through your notes!"
                    )
            except ValueError:
                logger.error(f"Note {note_num} not found in notes_map keys: {note_numbers}")
                await query.edit_message_text(
                    text="❌ Error finding next note. Please try again."
                )
        
    except Exception as e:
        logger.error(f"Error in handle_note_action: {str(e)}", exc_info=True)
        await query.edit_message_text(
            text="❌ Error performing action. Please try again later."
        )

async def handle_notes_page(update: Update, context: ContextTypes.DEFAULT_TYPE, callback_data: str):
    """Handle notes pagination."""
    try:
        query = update.callback_query
        chat_id = str(query.message.chat_id)
        
        # Parse callback data: notes_page_{page_num}
        parts = callback_data.split("_")
        if len(parts) < 3:
            logger.error(f"Invalid notes page callback data: {callback_data}")
            return
        
        page = int(parts[2])
        
        # Get notes map from bot storage
        bot = context.bot
        notes_map = getattr(bot, '_notes_storage', {}).get(chat_id, {})
        
        if not notes_map:
            await query.edit_message_text(
                text="❌ No notes found. Please try again later."
            )
            return
        
        # Create new keyboard with the requested page
        keyboard = create_notes_buttons_from_storage(notes_map, page)
        
        # Get original message text (before the "Tap a number" part)
        original_text = query.message.text
        if "\n\n📱 Tap a number to view that note:" in original_text:
            original_text = original_text.split("\n\n📱 Tap a number to view that note:")[0]
        
        await query.edit_message_text(
            text=f"{original_text}\n\n📱 Tap a number to view that note:",
            reply_markup=keyboard
        )
        
    except Exception as e:
        logger.error(f"Error in handle_notes_page: {str(e)}", exc_info=True)
        await query.edit_message_text(
            text="❌ Error loading page. Please try again later."
        )


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