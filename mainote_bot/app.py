"""
App module - Simplified version without database dependencies.
This module is kept for compatibility but doesn't contain any functionality.
"""
from mainote_bot.utils.logging import logger


async def start_webhook():
    """Stub function - webhook starting is handled in main.py."""
    logger.info("Webhook start requested but handled elsewhere in simplified bot")
    pass 