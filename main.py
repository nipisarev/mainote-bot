import asyncio
import logging
from fastapi import FastAPI
from telegram.ext import Application
from mainote_bot.config import TELEGRAM_BOT_TOKEN, WEBHOOK_URL
from mainote_bot.bot.handlers import setup_handlers
from mainote_bot.bot.webhook import create_app
from mainote_bot.utils.logging import setup_logging
from mainote_bot.database import init_pool, close_pool

# Setup logging
setup_logging()
logger = logging.getLogger(__name__)

# Create FastAPI app
app = create_app()

@app.on_event("startup")
async def startup_event():
    """Initialize the application on startup."""
    try:
        # Initialize database pool
        await init_pool()
        logger.info("Database pool initialized")

        # Initialize bot
        application = Application.builder().token(TELEGRAM_BOT_TOKEN).build()
        setup_handlers(application)
        
        # Set webhook
        await application.bot.set_webhook(url=WEBHOOK_URL)
        logger.info(f"Webhook set to {WEBHOOK_URL}")
        
        # Store application in app state
        app.state.application = application

        # Start the notification scheduler
        from mainote_bot.scheduler.notifications import start_scheduler
        start_scheduler()
        logger.info("Notification scheduler started")
        
        logger.info("Application initialized successfully")
    except Exception as e:
        logger.error(f"Failed to initialize application: {str(e)}", exc_info=True)
        raise

@app.on_event("shutdown")
async def shutdown_event():
    """Clean up resources on shutdown."""
    try:
        # Close database pool
        await close_pool()
        logger.info("Database pool closed")
        
        # Remove webhook
        if hasattr(app.state, 'application'):
            await app.state.application.bot.delete_webhook()
            logger.info("Webhook removed")
    except Exception as e:
        logger.error(f"Error during shutdown: {str(e)}", exc_info=True) 