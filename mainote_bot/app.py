from mainote_bot.database import init_db

async def start_webhook():
    """Start the webhook server."""
    try:
        # Start the scheduler
        start_scheduler()
        logger.info("Scheduler started successfully")
    except Exception as e:
        logger.error(f"Failed to start webhook: {e}")
        raise 