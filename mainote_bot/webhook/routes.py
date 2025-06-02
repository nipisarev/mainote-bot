import asyncio
from fastapi import APIRouter, Request, HTTPException
from telegram import Update
from mainote_bot.utils.logging import logger

router = APIRouter()

@router.post("/webhook")
async def webhook(request: Request):
    """Handle incoming webhook updates from Telegram."""
    try:
        # Get the bot and application from the app state
        bot = request.app.state.bot
        application = request.app.state.application

        if not bot or not application:
            raise HTTPException(status_code=500, detail="Bot not initialized")

        # Parse the update
        update_data = await request.json()
        update = Update.de_json(update_data, bot)
        logger.info(f"Received update: {update}")

        # Process update using the application
        await application.process_update(update)

        logger.info("Update processed successfully")
        return {"status": "OK"}

    except Exception as e:
        logger.error(f"Error processing webhook: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "OK"}

def create_app(bot, application):
    """Create and configure the FastAPI router."""
    return router
