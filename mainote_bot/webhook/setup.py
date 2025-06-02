import asyncio
from mainote_bot.config import WEBHOOK_URL
from mainote_bot.utils.logging import logger

async def verify_webhook(webhook_url, webhook_info):
    """Verify that the webhook is set up correctly."""
    if not webhook_info.url:
        logger.error("Webhook URL is empty after setting")
        return False

    if webhook_info.url != webhook_url:
        logger.error(f"Webhook URL mismatch. Expected: {webhook_url}, Got: {webhook_info.url}")
        return False

    logger.info("Webhook setup successful")
    return True

async def handle_retry(attempt, max_retries, retry_delay):
    """Handle retry logic for webhook setup."""
    if attempt < max_retries - 1:
        logger.info(f"Retrying in {retry_delay} seconds...")
        await asyncio.sleep(retry_delay)
        return True
    return False

async def setup_webhook(bot):
    """Set up the webhook and verify it's working."""
    if not WEBHOOK_URL:
        logger.error("WEBHOOK_URL is not set")
        return False

    webhook_url = f"{WEBHOOK_URL}/webhook"
    logger.info(f"Setting up webhook with URL: {webhook_url}")

    max_retries = 3
    retry_delay = 5  # seconds

    for attempt in range(max_retries):
        try:
            # Get current webhook info
            current_webhook = await bot.get_webhook_info()
            logger.info(f"Current webhook info: {current_webhook}")

            # Remove any existing webhook
            await bot.delete_webhook()
            logger.info("Removed existing webhook")

            # Set new webhook
            await bot.set_webhook(url=webhook_url)
            logger.info(f"Set webhook to: {webhook_url}")

            # Verify webhook
            webhook_info = await bot.get_webhook_info()
            logger.info(f"New webhook info: {webhook_info}")

            if await verify_webhook(webhook_url, webhook_info):
                return True

            # If verification failed, try to retry
            if await handle_retry(attempt, max_retries, retry_delay):
                continue
            return False

        except Exception as e:
            logger.error(f"Error setting up webhook (attempt {attempt + 1}/{max_retries}): {str(e)}", exc_info=True)
            if await handle_retry(attempt, max_retries, retry_delay):
                continue
            return False
    return None