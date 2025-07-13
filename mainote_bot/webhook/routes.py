import asyncio
import os
import httpx
from datetime import datetime
from fastapi import APIRouter, Request, HTTPException, Header
from telegram import Update
from mainote_bot.utils.logging import logger
from mainote_bot.config import INTERNAL_API_KEY
from pydantic import BaseModel
from typing import Optional

# Add Pydantic model for notification request
class NotificationRequest(BaseModel):
    chat_id: str
    type: str  # morning_notification/ai_summary/reminder
    message: str
    id: str  # UUID for logging

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

@router.post("/notification")
async def notification(
    request: Request, 
    notification_data: NotificationRequest,
    x_internal_api_key: str = Header(None, alias="X-Internal-API-Key")
):
    """Handle internal notification requests from mainote_server."""
    try:
        # Check internal API key for authorization
        if x_internal_api_key != INTERNAL_API_KEY:
            logger.warning(f"Unauthorized notification request - Invalid API key from {request.client.host}")
            raise HTTPException(status_code=401, detail="Unauthorized: Invalid internal API key")
        
        # Get the bot from the app state
        bot = request.app.state.bot
        
        if not bot:
            raise HTTPException(status_code=500, detail="Bot not initialized")
        
        # Extract notification data
        chat_id = notification_data.chat_id
        notification_type = notification_data.type
        message = notification_data.message
        notification_id = notification_data.id
        
        # Log the notification request
        logger.info(f"Received notification request - ID: {notification_id}, Type: {notification_type}, Chat: {chat_id}")
        
        # Send message to the user
        await bot.send_message(
            chat_id=chat_id,
            text=message,
            parse_mode=None  # Send as plain text to avoid formatting issues
        )
        
        logger.info(f"Successfully sent notification {notification_id} to chat {chat_id}")
        
        return {
            "status": "success",
            "message": "Notification sent successfully",
            "notification_id": notification_id,
            "chat_id": chat_id,
            "type": notification_type
        }
        
    except Exception as e:
        logger.error(f"Error processing notification {notification_data.id}: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Failed to send notification: {str(e)}")

@router.get("/health")
async def health(request: Request):
    """Comprehensive health check endpoint for all services."""
    health_status = {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "services": {
            "python_bot": {"status": "unknown"},
            "mainote_server": {"status": "unknown"}
        }
    }
    
    overall_healthy = True
    
    # Check Python bot service
    try:
        bot = getattr(request.app.state, 'bot', None)
        application = getattr(request.app.state, 'application', None)
        
        if bot and application:
            health_status["services"]["python_bot"] = {
                "status": "healthy",
                "bot_initialized": True,
                "application_initialized": True
            }
        else:
            health_status["services"]["python_bot"] = {
                "status": "unhealthy",
                "error": "Bot or application not initialized",
                "bot_initialized": bot is not None,
                "application_initialized": application is not None
            }
            overall_healthy = False
    except Exception as e:
        health_status["services"]["python_bot"] = {
            "status": "unhealthy",
            "error": str(e)
        }
        overall_healthy = False
    
    # Check Mainote server service
    try:
        mainote_server_url = os.getenv('MAINOTE_SERVER_URL', 'http://mainote-server:8081')
        timeout = httpx.Timeout(5.0)  # 5 second timeout
        
        async with httpx.AsyncClient(timeout=timeout) as client:
            response = await client.get(f"{mainote_server_url}/health")
            
            if response.status_code == 200:
                server_health_data = response.json()
                health_status["services"]["mainote_server"] = {
                    "status": "healthy",
                    "response_time_ms": response.elapsed.total_seconds() * 1000,
                    "backend_status": server_health_data.get("status", "unknown"),
                    "backend_version": server_health_data.get("version", "unknown")
                }
            else:
                health_status["services"]["mainote_server"] = {
                    "status": "unhealthy",
                    "error": f"HTTP {response.status_code}",
                    "url": mainote_server_url
                }
                overall_healthy = False
    except httpx.TimeoutException:
        health_status["services"]["mainote_server"] = {
            "status": "unhealthy",
            "error": "Connection timeout",
            "url": os.getenv('MAINOTE_SERVER_URL', 'http://mainote-server:8081')
        }
        overall_healthy = False
    except Exception as e:
        health_status["services"]["mainote_server"] = {
            "status": "unhealthy",
            "error": str(e),
            "url": os.getenv('MAINOTE_SERVER_URL', 'http://mainote-server:8081')
        }
        overall_healthy = False
    
    # Set overall status
    if not overall_healthy:
        health_status["status"] = "degraded"
        
    # Count healthy services
    healthy_services = sum(1 for service in health_status["services"].values() 
                          if service.get("status") == "healthy")
    total_services = len(health_status["services"])
    
    health_status["summary"] = {
        "healthy_services": healthy_services,
        "total_services": total_services,
        "health_percentage": round((healthy_services / total_services) * 100, 1)
    }
    
    # Return appropriate HTTP status code
    if overall_healthy:
        return health_status
    else:
        logger.warning(f"Health check failed: {health_status}")
        raise HTTPException(
            status_code=503,
            detail=health_status
        )

def create_app(bot, application):
    """Create and configure the FastAPI router."""
    return router
