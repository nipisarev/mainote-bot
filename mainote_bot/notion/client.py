from notion_client import Client
from mainote_bot.config import NOTION_API_KEY
from mainote_bot.utils.logging import logger

def create_notion_client():
    """Create and return a Notion client instance."""
    try:
        client = Client(auth=NOTION_API_KEY)
        logger.info("Notion client created successfully")
        return client
    except Exception as e:
        logger.error(f"Error creating Notion client: {str(e)}", exc_info=True)
        raise

# Create a global Notion client instance
notion_client = None

def get_notion_client():
    """Get the global Notion client instance, creating it if necessary."""
    global notion_client
    if notion_client is None:
        notion_client = create_notion_client()
    return notion_client