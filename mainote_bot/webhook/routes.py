import asyncio
from flask import Flask, request
from telegram import Update
from mainote_bot.utils.logging import logger

def webhook(bot, application):
    """Handle incoming webhook updates from Telegram."""
    if not bot or not application:
        return 'Error: Bot not initialized', 500

    try:
        update = Update.de_json(request.get_json(), bot)
        logger.info(f"Received update: {update}")

        # Create a new event loop for this request
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)

        # Make sure the bot is initialized in this event loop
        loop.run_until_complete(bot.initialize())

        # Re-initialize the application in this event loop
        loop.run_until_complete(application.initialize())

        # Process update using the application
        loop.run_until_complete(application.process_update(update))

        # Wait for any pending tasks to complete
        pending = asyncio.all_tasks(loop)
        if pending:
            loop.run_until_complete(asyncio.gather(*pending, return_exceptions=True))

        logger.info("Update processed successfully")

        return 'OK'
    except Exception as e:
        logger.error(f"Error processing webhook: {str(e)}", exc_info=True)
        return 'Error', 500

def create_app(bot, application):
    """Create and configure the Flask app."""
    # Create Flask app
    app = Flask(__name__)

    # Define health check endpoint
    @app.route('/health', methods=['GET'])
    def health():
        """Health check endpoint."""
        return 'OK'

    # Create a closure that includes bot and application
    def webhook_handler():
        return webhook(bot, application)

    # Register the routes
    app.add_url_rule('/webhook', 'webhook', webhook_handler, methods=['POST'])

    return app
