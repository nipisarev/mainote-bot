from mainote_bot.notion.client import get_notion_client
from mainote_bot.config import NOTION_DATABASE_ID, NOTE_CATEGORIES
from mainote_bot.utils.logging import logger

async def get_active_tasks():
    """Query Notion database for active tasks."""
    try:
        notion = get_notion_client()
        # Query Notion for active tasks
        response = notion.databases.query(
            database_id=NOTION_DATABASE_ID,
            filter={
                "and": [
                    {
                        "property": "Status",
                        "select": {
                            "equals": "active"
                        }
                    }
                ]
            }
        )

        logger.info(f"Found {len(response['results'])} active tasks in Notion")
        return response["results"]
    except Exception as e:
        logger.error(f"Error querying Notion for active tasks: {str(e)}", exc_info=True)
        return []

async def format_morning_notification(tasks):
    """Format the morning notification message."""
    if not tasks:
        return "Доброе утро! У вас нет активных задач на сегодня. Хорошего дня! 🌞"

    message = "🌅 Доброе утро! Вот ваш план на сегодня:\n\n"

    for i, task in enumerate(tasks, 1):
        try:
            # Extract task details
            title = task["properties"]["Name"]["title"][0]["text"]["content"] if task["properties"]["Name"]["title"] else "Без названия"
            note_type = task["properties"]["Type"]["select"]["name"] if "select" in task["properties"]["Type"] else "task"

            # Get emoji for task type
            emoji = {
                "idea": "💡",
                "task": "✅",
                "personal": "🏖"
            }.get(note_type, "📝")

            message += f"{i}. {emoji} {title}\n"
        except Exception as e:
            logger.error(f"Error formatting task {i}: {str(e)}", exc_info=True)
            continue

    message += "\nУдачного и продуктивного дня! 💪"
    return message

def create_note(text):
    """Create a new note in Notion."""
    try:
        notion = get_notion_client()
        new_page = {
            "parent": {"database_id": NOTION_DATABASE_ID},
            "properties": {
                "Name": {
                    "title": [
                        {
                            "text": {
                                "content": text[:50] + "..." if len(text) > 50 else text
                            }
                        }
                    ]
                },
                "Type": {
                    "select": {
                        "name": "task"
                    }
                },
                "Status": {
                    "select": {
                        "name": "active"
                    }
                },
                "Source": {
                    "rich_text": [
                        {
                            "text": {
                                "content": "telegram-text"
                            }
                        }
                    ]
                },
                "Content": {
                    "rich_text": [
                        {
                            "text": {
                                "content": text
                            }
                        }
                    ]
                }
            }
        }

        page = notion.pages.create(**new_page)
        logger.info(f"Saved to Notion with ID: {page['id']}")
        return page
    except Exception as e:
        logger.error(f"Error creating note in Notion: {str(e)}", exc_info=True)
        raise

def update_note_type(page_id, note_type):
    """Update the type of a note in Notion."""
    try:
        notion = get_notion_client()
        notion.pages.update(
            page_id=page_id,
            properties={
                "Type": {
                    "select": {
                        "name": note_type
                    }
                }
            }
        )
        logger.info(f"Updated note type to {note_type} for page {page_id}")
        return True
    except Exception as e:
        logger.error(f"Error updating note type: {str(e)}", exc_info=True)
        return False