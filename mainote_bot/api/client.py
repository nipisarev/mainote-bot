import httpx
import json
from typing import Dict, Any, Optional, List
from mainote_bot.config import SERVER_URL
from mainote_bot.utils.logging import logger


class MainoteAPIClient:
    """Client for interacting with the Mainote backend API."""
    
    def __init__(self, server_url: str = SERVER_URL):
        self.server_url = server_url.rstrip('/')
        
    async def health_check(self) -> bool:
        """Check if the backend server is healthy."""
        url = f"{self.server_url}/health"
        
        try:
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.get(url)
                return response.status_code == 200
        except Exception as e:
            logger.error(f"Health check failed: {str(e)}")
            return False

    async def get_user_by_chat_id(self, chat_id: str) -> Optional[Dict[str, Any]]:
        """
        Check if a user exists for the given chat_id.
        
        Args:
            chat_id: Telegram chat ID as string
            
        Returns:
            User data if found, None if not found
            
        Raises:
            Exception: If API call fails
        """
        try:
            url = f"{self.server_url}/api/v1/users/chat/{chat_id}"
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.get(url)
                
                if response.status_code == 200:
                    user_data = response.json()
                    logger.info(f"Found existing user for chat_id {chat_id}: {user_data.get('email')}")
                    return user_data
                elif response.status_code == 404:
                    logger.info(f"No user found for chat_id {chat_id}")
                    return None
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error checking user by chat_id {chat_id}: {response.status_code} - {error_data}")
                    raise Exception(f"API error: {response.status_code} - {error_data.get('message', 'Unknown error')}")
                    
        except httpx.TimeoutException:
            logger.error(f"Timeout checking user by chat_id {chat_id}")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error checking user by chat_id {chat_id}: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if "API error" in str(e) or "Server timeout" in str(e) or "Unable to connect" in str(e):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error checking user by chat_id {chat_id}: {str(e)}")
            raise
    
    async def create_user(self, email: str, password: str) -> Dict[str, Any]:
        """
        Create a new user.
        
        Args:
            email: User's email address
            password: User's password
            
        Returns:
            Created user data
            
        Raises:
            Exception: If user creation fails
        """
        try:
            url = f"{self.server_url}/api/v1/users"
            payload = {
                "email": email,
                "password": password
            }
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(
                    url,
                    json=payload,
                    headers={"Content-Type": "application/json"}
                )
                
                if response.status_code == 201:
                    user_data = response.json()
                    logger.info(f"Created new user: {email}")
                    return user_data
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error creating user {email}: {response.status_code} - {error_data}")
                    
                    if response.status_code == 409:
                        raise Exception("User with this email already exists")
                    elif response.status_code == 400:
                        raise Exception(f"Invalid input: {error_data.get('message', 'Please check your email and password')}")
                    else:
                        raise Exception(f"Unable to create user: {error_data.get('message', 'Unknown error')}")
                        
        except httpx.TimeoutException:
            logger.error(f"Timeout creating user {email}")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error creating user {email}: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if any(msg in str(e) for msg in ["User with this email already exists", "Invalid input", "Unable to create user", "Server timeout", "Unable to connect"]):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error creating user {email}: {str(e)}")
            raise Exception("An unexpected error occurred - please try again later")
    
    async def auth_user(self, email: str, password: str, chat_id: str) -> Optional[Dict[str, Any]]:
        """Authenticate user and associate with chat_id."""
        url = f"{self.server_url}/api/v1/auth"
        payload = {
            "email": email,
            "password": password,
            "chat_id": chat_id
        }
        
        try:
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(
                    url,
                    json=payload,
                    headers={'Content-Type': 'application/json'}
                )
                if response.status_code in [200, 201]:
                    result = response.json()
                    logger.info(f"Successfully authenticated user for chat_id {chat_id}")
                    return result
                else:
                    logger.error(f"Failed to authenticate user: {response.status_code} - {response.text}")
                    return None
        except Exception as e:
            logger.error(f"Error authenticating user: {str(e)}", exc_info=True)
            return None
    
    async def authenticate_user(self, chat_id: str, email: str, password: str) -> Dict[str, Any]:
        """
        Authenticate user and associate with chat_id.
        This endpoint either authenticates existing user or creates new user and links to chat_id.
        
        Args:
            chat_id: Telegram chat ID as string
            email: User's email address
            password: User's password
            
        Returns:
            Authentication response with user data
            
        Raises:
            Exception: If authentication fails
        """
        try:
            url = f"{self.server_url}/api/v1/auth"
            payload = {
                "chat_id": chat_id,
                "email": email,
                "password": password
            }
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(
                    url,
                    json=payload,
                    headers={"Content-Type": "application/json"}
                )
                
                if response.status_code in [200, 201]:
                    auth_data = response.json()
                    action = "authenticated" if response.status_code == 200 else "created and authenticated"
                    logger.info(f"User {email} {action} and linked to chat_id {chat_id}")
                    return auth_data
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error authenticating user {email} with chat_id {chat_id}: {response.status_code} - {error_data}")
                    
                    if response.status_code == 401:
                        raise Exception("Invalid email or password")
                    elif response.status_code == 409:
                        raise Exception("This chat is already connected to another user")
                    elif response.status_code == 400:
                        raise Exception(f"Invalid input: {error_data.get('message', 'Please check your credentials')}")
                    else:
                        raise Exception(f"Authentication failed: {error_data.get('message', 'Unknown error')}")
                        
        except httpx.TimeoutException:
            logger.error(f"Timeout authenticating user {email} with chat_id {chat_id}")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error authenticating user {email} with chat_id {chat_id}: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if any(msg in str(e) for msg in ["Invalid email", "already connected", "Invalid input", "Authentication failed", "Server timeout", "Unable to connect"]):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error authenticating user {email} with chat_id {chat_id}: {str(e)}")
            raise Exception("An unexpected error occurred - please try again later")

    async def create_user_settings(self, user_id: str, chat_id: str, morning_notification_time: str = None, timezone: str = None) -> Dict[str, Any]:
        """
        Create user settings for a given user_id and chat_id.
        
        Args:
            user_id: User UUID as string
            chat_id: Telegram chat ID as string
            morning_notification_time: Optional morning notification time (HH:MM format)
            timezone: Optional timezone (e.g., "UTC", "America/New_York")
            
        Returns:
            Created user settings data
            
        Raises:
            Exception: If creation fails
        """
        try:
            url = f"{self.server_url}/api/v1/user/settings"
            payload = {
                "user_id": user_id,
                "chat_id": chat_id
            }
            
            # Add optional fields if provided
            if morning_notification_time:
                payload["morning_notification_time"] = morning_notification_time
            if timezone:
                payload["timezone"] = timezone
            
            logger.info(f"Creating user settings with payload: {payload}")  # Debug log
            
            logger.info(f"Creating user settings with payload: {payload}")  # Debug log
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(
                    url,
                    json=payload,
                    headers={"Content-Type": "application/json"}
                )
                
                if response.status_code == 201:
                    settings_data = response.json()
                    logger.info(f"Successfully created user settings for chat_id {chat_id}")
                    return settings_data
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error creating user settings for chat_id {chat_id}: {response.status_code} - {error_data}")
                    
                    if response.status_code == 400:
                        raise Exception(f"Invalid input: {error_data.get('message', 'Please check your data')}")
                    elif response.status_code == 404:
                        raise Exception("User not found")
                    else:
                        raise Exception(f"Settings creation failed: {error_data.get('message', 'Unknown error')}")
                        
        except httpx.TimeoutException:
            logger.error(f"Timeout creating user settings for chat_id {chat_id}")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error creating user settings for chat_id {chat_id}: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if any(msg in str(e) for msg in ["Invalid input", "User not found", "Settings creation failed", "Server timeout", "Unable to connect"]):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error creating user settings for chat_id {chat_id}: {str(e)}")
            raise Exception("An unexpected error occurred - please try again later")
