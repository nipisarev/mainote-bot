import os
import asyncio
from asyncpg import create_pool
from mainote_bot.utils.logging import logger

# Global pool instance
_pool = None
_pool_lock = asyncio.Lock()

async def init_pool():
    """Initialize the database connection pool."""
    global _pool
    if _pool is None:
        async with _pool_lock:
            if _pool is None:  # Double-check pattern
                try:
                    database_url = os.getenv('DATABASE_URL')
                    if not database_url:
                        raise ValueError("DATABASE_URL environment variable is not set")
                    
                    # Ensure the URL uses postgresql:// instead of postgres://
                    if database_url.startswith('postgres://'):
                        database_url = database_url.replace('postgres://', 'postgresql://', 1)
                    
                    _pool = await create_pool(
                        dsn=database_url,
                        min_size=1,
                        max_size=10,
                        command_timeout=60
                    )
                    logger.info("Database connection pool initialized")
                except Exception as e:
                    logger.error(f"Failed to initialize database pool: {str(e)}", exc_info=True)
                    raise
    return _pool

async def get_pool():
    """Get the database connection pool, initializing it if necessary."""
    if _pool is None:
        await init_pool()
    return _pool

async def close_pool():
    """Close the database connection pool."""
    global _pool
    if _pool is not None:
        await _pool.close()
        _pool = None
        logger.info("Database connection pool closed")

async def get_user_preferences(chat_id):
    """Get user preferences from database."""
    try:
        pool = await get_pool()
        async with pool.acquire() as conn:
            row = await conn.fetchrow(
                'SELECT notification_time, timezone FROM user_preferences WHERE chat_id = $1',
                str(chat_id)
            )
            return dict(row) if row else {}
    except Exception as e:
        logger.error(f"Error getting user preferences: {str(e)}", exc_info=True)
        return {}

async def set_user_preferences(chat_id, preferences):
    """Set user preferences in database."""
    try:
        pool = await get_pool()
        async with pool.acquire() as conn:
            await conn.execute('''
                INSERT INTO user_preferences (chat_id, notification_time, timezone, updated_at)
                VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
                ON CONFLICT (chat_id) DO UPDATE
                SET notification_time = $2,
                    timezone = $3,
                    updated_at = CURRENT_TIMESTAMP
            ''', str(chat_id), preferences.get('notification_time'), preferences.get('timezone'))
            return True
    except Exception as e:
        logger.error(f"Error setting user preferences: {str(e)}", exc_info=True)
        return False

async def get_all_users_with_preferences():
    """Get all users who have set preferences."""
    try:
        pool = await get_pool()
        async with pool.acquire() as conn:
            rows = await conn.fetch('SELECT chat_id FROM user_preferences')
            return [row['chat_id'] for row in rows]
    except Exception as e:
        logger.error(f"Error getting all users with preferences: {str(e)}", exc_info=True)
        return [] 