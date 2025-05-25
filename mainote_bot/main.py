import sys
import asyncio
from telegram import Bot
from telegram.ext import Application, CommandHandler, MessageHandler, CallbackQueryHandler, filters
from gunicorn.app.base import BaseApplication
from mainote_bot.config import TELEGRAM_BOT_TOKEN, NOTION_API_KEY
from mainote_bot.utils.logging import logger
from mainote_bot.bot.commands import start_command, help_command, morning_command, settime_command, settimezone_command
from mainote_bot.bot.messages import handle_message
from mainote_bot.bot.callbacks import button_callback
from mainote_bot.scheduler.notifications import start_scheduler
from mainote_bot.webhook.setup import setup_webhook
from mainote_bot.webhook.routes import create_app

def initialize_bot():
    """Initialize the bot and application."""
    try:
        # Create bot instance
        bot = Bot(token=TELEGRAM_BOT_TOKEN)

        # Create application instance
        application = Application.builder().token(TELEGRAM_BOT_TOKEN).build()

        # Add handlers
        application.add_handler(CommandHandler("start", start_command))
        application.add_handler(CommandHandler("help", help_command))
        application.add_handler(CommandHandler("morning", morning_command))
        application.add_handler(CommandHandler("settime", settime_command))
        application.add_handler(CommandHandler("settimezone", settimezone_command))
        application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message))
        application.add_handler(CallbackQueryHandler(button_callback))

        # Initialize the application with a new event loop
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        loop.run_until_complete(application.initialize())
        # Don't close the loop here as it will be used for future operations

        logger.info("Successfully initialized bot and notion client")
        return bot, application, True
    except Exception as e:
        logger.error(f"Failed to initialize bot or notion client: {str(e)}", exc_info=True)
        return None, None, False

class FlaskApplication(BaseApplication):
    """Gunicorn application for running Flask."""
    def __init__(self, app, options=None):
        self.options = options or {}
        self.application = app
        super().__init__()

    def load_config(self):
        for key, value in self.options.items():
            self.cfg.set(key, value)

    def load(self):
        return self.application

def main():
    """Main entry point for the application."""
    # Initialize bot
    bot, application, success = initialize_bot()
    if not success:
        logger.error("Failed to initialize bot, exiting...")
        sys.exit(1)

    # Set up webhook using the current event loop
    loop = asyncio.get_event_loop()
    if not loop.run_until_complete(setup_webhook(bot)):
        logger.error("Failed to set up webhook, but continuing anyway...")

    # Start the scheduler
    start_scheduler()

    # Create Flask app
    app = create_app(bot, application)

    # Run Flask app with gunicorn
    options = {
        'bind': '0.0.0.0:8080',
        'workers': 1,
        'worker_class': 'sync',
        'timeout': 120,
        'accesslog': '-',
        'errorlog': '-',
        'loglevel': 'info',
        'worker_connections': 1000
    }

    FlaskApplication(app, options).run()

if __name__ == '__main__':
    main()