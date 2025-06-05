import asyncio
import os
import httpx
from datetime import datetime
from fastapi import APIRouter, Request, HTTPException
from telegram import Update
from mainote_bot.utils.logging import logger
from mainote_bot.database import get_pool

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
async def health(request: Request):
    """Comprehensive health check endpoint for all services."""
    health_status = {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "services": {
            "python_bot": {"status": "unknown"},
            "database": {"status": "unknown"},
            "go_backend": {"status": "unknown"}
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
    
    # Check PostgreSQL database
    try:
        pool = await get_pool()
        async with pool.acquire() as conn:
            # Simple query to test database connectivity
            result = await conn.fetchval('SELECT 1')
            if result == 1:
                health_status["services"]["database"] = {
                    "status": "healthy",
                    "type": "postgresql",
                    "connection": "active"
                }
            else:
                health_status["services"]["database"] = {
                    "status": "unhealthy",
                    "error": "Database query returned unexpected result"
                }
                overall_healthy = False
    except Exception as e:
        health_status["services"]["database"] = {
            "status": "unhealthy",
            "error": str(e),
            "type": "postgresql"
        }
        overall_healthy = False
    
    # Check Go backend service
    try:
        go_backend_url = os.getenv('GO_BACKEND_URL', 'http://go-backend:8081')
        timeout = httpx.Timeout(5.0)  # 5 second timeout
        
        async with httpx.AsyncClient(timeout=timeout) as client:
            response = await client.get(f"{go_backend_url}/health")
            
            if response.status_code == 200:
                go_health_data = response.json()
                health_status["services"]["go_backend"] = {
                    "status": "healthy",
                    "response_time_ms": response.elapsed.total_seconds() * 1000,
                    "backend_status": go_health_data.get("status", "unknown"),
                    "backend_version": go_health_data.get("version", "unknown")
                }
            else:
                health_status["services"]["go_backend"] = {
                    "status": "unhealthy",
                    "error": f"HTTP {response.status_code}",
                    "url": go_backend_url
                }
                overall_healthy = False
    except httpx.TimeoutException:
        health_status["services"]["go_backend"] = {
            "status": "unhealthy",
            "error": "Connection timeout",
            "url": os.getenv('GO_BACKEND_URL', 'http://go-backend:8081')
        }
        overall_healthy = False
    except Exception as e:
        health_status["services"]["go_backend"] = {
            "status": "unhealthy",
            "error": str(e),
            "url": os.getenv('GO_BACKEND_URL', 'http://go-backend:8081')
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
