"""
Database module - Simplified version without actual database connections.
This module provides stub functions for compatibility but doesn't connect to any database.
"""
from mainote_bot.utils.logging import logger


async def get_pool():
    """Stub function - returns None as we don't use database connections."""
    logger.info("Database pool requested but not implemented in simplified bot")
    return None


async def init_pool():
    """Stub function - no database initialization needed."""
    logger.info("Database initialization skipped in simplified bot")
    return None


async def close_pool():
    """Stub function - no database connections to close."""
    logger.info("Database pool close requested but not implemented in simplified bot")
    return None 