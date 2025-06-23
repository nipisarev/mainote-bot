import sys
from telegram import Bot
from telegram.ext import Application, CommandHandler, MessageHandler, CallbackQueryHandler, filters
from fastapi import FastAPI
from mainote_bot.config import TELEGRAM_BOT_TOKEN
from mainote_bot.utils.logging import logger
from mainote_bot.bot.commands import start_command, help_command, reset_command
from mainote_bot.bot.messages import handle_message
from mainote_bot.bot.callbacks import button_callback
from mainote_bot.webhook.setup import setup_webhook
from mainote_bot.webhook.routes import create_app

# Create FastAPI app
app = FastAPI()

async def initialize_bot():
    """Initialize the bot and application."""
    try:
        # Create bot instance
        bot = Bot(token=TELEGRAM_BOT_TOKEN)
        await bot.initialize()

        # Create application instance
        application = Application.builder().token(TELEGRAM_BOT_TOKEN).build()

        # Add handlers
        application.add_handler(CommandHandler("start", start_command))
        application.add_handler(CommandHandler("help", help_command))
        application.add_handler(CommandHandler("reset", reset_command))
        application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message))
        application.add_handler(CallbackQueryHandler(button_callback))

        # Initialize the application
        await application.initialize()
        await application.start()

        logger.info("Successfully initialized bot")
        return bot, application, True
    except Exception as e:
        logger.error(f"Failed to initialize bot: {str(e)}", exc_info=True)
        return None, None, False

@app.on_event("startup")
async def startup_event():
    """Initialize bot on startup."""
    # Initialize bot
    bot, application, success = await initialize_bot()
    if not success:
        logger.error("Failed to initialize bot, exiting...")
        sys.exit(1)

    # Set up webhook
    if not await setup_webhook(bot):
        logger.error("Failed to set up webhook, but continuing anyway...")

    # Start the scheduler - disabled for simplified bot
    # start_scheduler()

    # Store bot and application in app state
    app.state.bot = bot
    app.state.application = application

@app.on_event("shutdown")
async def shutdown_event():
    """Clean up resources on shutdown."""
    if hasattr(app.state, 'application'):
        await app.state.application.stop()
        await app.state.application.shutdown()

# Include webhook routes
app.include_router(create_app(None, None))

if __name__ == '__main__':
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)