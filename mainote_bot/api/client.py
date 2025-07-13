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
    
    # Note-related methods
    async def create_note(self, chat_id: str, content: str, title: str = None, category: str = "idea", 
                         status: str = "active", source: str = "telegram", 
                         voice_file_id: str = None, transcription: str = None, 
                         metadata: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Create a new note.
        
        Args:
            chat_id: Telegram chat ID as string
            content: Note content (required)
            title: Note title (optional)
            category: Note category (default: "idea")
            status: Note status (default: "active")
            source: Note source (default: "telegram")
            voice_file_id: Telegram voice file ID (optional)
            transcription: Voice transcription (optional)
            metadata: Additional metadata (optional)
            
        Returns:
            Created note data
            
        Raises:
            Exception: If note creation fails
        """
        try:
            url = f"{self.server_url}/api/v1/notes"
            payload = {
                "chat_id": chat_id,
                "content": content,
                "category": category,
                "status": status,
                "source": source
            }
            
            # Add optional fields if provided
            if title:
                payload["title"] = title
            if voice_file_id:
                payload["voice_file_id"] = voice_file_id
            if transcription:
                payload["transcription"] = transcription
            if metadata:
                payload["metadata"] = metadata
            
            logger.info(f"Creating note for chat_id {chat_id} with category {category}")
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(
                    url,
                    json=payload,
                    headers={"Content-Type": "application/json"}
                )
                
                if response.status_code == 201:
                    note_data = response.json()
                    logger.info(f"Successfully created note {note_data.get('note_id')} for chat_id {chat_id}")
                    return note_data
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error creating note for chat_id {chat_id}: {response.status_code} - {error_data}")
                    
                    if response.status_code == 400:
                        raise Exception(f"Invalid input: {error_data.get('message', 'Please check your data')}")
                    elif response.status_code == 404:
                        raise Exception("User not found - please complete setup first")
                    else:
                        raise Exception(f"Note creation failed: {error_data.get('message', 'Unknown error')}")
                        
        except httpx.TimeoutException:
            logger.error(f"Timeout creating note for chat_id {chat_id}")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error creating note for chat_id {chat_id}: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if any(msg in str(e) for msg in ["Invalid input", "User not found", "Note creation failed", "Server timeout", "Unable to connect"]):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error creating note for chat_id {chat_id}: {str(e)}")
            raise Exception("An unexpected error occurred - please try again later")

    async def get_notes(self, chat_id: str, category: str = None, status: str = "active", 
                       limit: int = 50, offset: int = 0) -> Dict[str, Any]:
        """
        Get notes for a user.
        
        Args:
            chat_id: Telegram chat ID as string
            category: Filter by category (optional)
            status: Filter by status (default: "active")
            limit: Maximum number of notes to return (default: 50)
            offset: Number of notes to skip (default: 0)
            
        Returns:
            Notes list data with pagination info
            
        Raises:
            Exception: If getting notes fails
        """
        try:
            url = f"{self.server_url}/api/v1/notes"
            params = {
                "chat_id": chat_id,
                "status": status,
                "limit": limit,
                "offset": offset
            }
            
            if category:
                params["category"] = category
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.get(url, params=params)
                
                if response.status_code == 200:
                    notes_data = response.json()
                    logger.info(f"Retrieved {len(notes_data.get('notes', []))} notes for chat_id {chat_id}")
                    return notes_data
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error getting notes for chat_id {chat_id}: {response.status_code} - {error_data}")
                    
                    if response.status_code == 404:
                        raise Exception("User not found - please complete setup first")
                    else:
                        raise Exception(f"Failed to get notes: {error_data.get('message', 'Unknown error')}")
                        
        except httpx.TimeoutException:
            logger.error(f"Timeout getting notes for chat_id {chat_id}")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error getting notes for chat_id {chat_id}: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if any(msg in str(e) for msg in ["User not found", "Failed to get notes", "Server timeout", "Unable to connect"]):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error getting notes for chat_id {chat_id}: {str(e)}")
            raise Exception("An unexpected error occurred - please try again later")

    # Apps and Integration methods
    async def get_apps(self) -> List[Dict[str, Any]]:
        """
        Get all available apps for installation.
        
        Returns:
            List of available apps
            
        Raises:
            Exception: If getting apps fails
        """
        try:
            url = f"{self.server_url}/api/v1/apps"
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.get(url)
                
                if response.status_code == 200:
                    apps_data = response.json()
                    logger.info(f"Retrieved {len(apps_data.get('apps', []))} available apps")
                    return apps_data.get('apps', [])
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error getting apps: {response.status_code} - {error_data}")
                    raise Exception(f"Failed to get apps: {error_data.get('message', 'Unknown error')}")
                    
        except httpx.TimeoutException:
            logger.error("Timeout getting apps")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error getting apps: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if any(msg in str(e) for msg in ["Failed to get apps", "Server timeout", "Unable to connect"]):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error getting apps: {str(e)}")
            raise Exception("An unexpected error occurred - please try again later")

    async def get_integrations(self, chat_id: str, status: str = None, app_id: str = None, 
                              limit: int = 50, offset: int = 0) -> Dict[str, Any]:
        """
        Get integrations for a user.
        
        Args:
            chat_id: Telegram chat ID as string
            status: Filter by status (optional)
            app_id: Filter by app ID (optional)
            limit: Maximum number of integrations to return (default: 50)
            offset: Number of integrations to skip (default: 0)
            
        Returns:
            Integrations list data with pagination info
            
        Raises:
            Exception: If getting integrations fails
        """
        try:
            url = f"{self.server_url}/api/v1/integrations"
            params = {
                "chat_id": chat_id,
                "limit": limit,
                "offset": offset
            }
            
            if status:
                params["status"] = status
            if app_id:
                params["app_id"] = app_id
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.get(url, params=params)
                
                if response.status_code == 200:
                    integrations_data = response.json()
                    logger.info(f"Retrieved {len(integrations_data.get('integrations', []))} integrations for chat_id {chat_id}")
                    return integrations_data
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error getting integrations for chat_id {chat_id}: {response.status_code} - {error_data}")
                    
                    if response.status_code == 404:
                        raise Exception("User not found - please complete setup first")
                    else:
                        raise Exception(f"Failed to get integrations: {error_data.get('message', 'Unknown error')}")
                        
        except httpx.TimeoutException:
            logger.error(f"Timeout getting integrations for chat_id {chat_id}")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error getting integrations for chat_id {chat_id}: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if any(msg in str(e) for msg in ["User not found", "Failed to get integrations", "Server timeout", "Unable to connect"]):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error getting integrations for chat_id {chat_id}: {str(e)}")
            raise Exception("An unexpected error occurred - please try again later")

    async def create_integration(self, chat_id: str, app_id: str, auth_type: str, 
                               auth_data: Dict[str, Any], config: Dict[str, Any] = None,
                               status: str = "active") -> Dict[str, Any]:
        """
        Create a new integration.
        
        Args:
            chat_id: Telegram chat ID as string
            app_id: App ID to integrate with
            auth_type: Authentication type (e.g., "api_key", "oauth2")
            auth_data: Authentication data specific to the provider
            config: Integration configuration (optional)
            status: Integration status (default: "active")
            
        Returns:
            Created integration data
            
        Raises:
            Exception: If integration creation fails
        """
        try:
            url = f"{self.server_url}/api/v1/integrations"
            payload = {
                "chat_id": chat_id,
                "app_id": app_id,
                "auth_type": auth_type,
                "auth_data": auth_data,
                "status": status
            }
            
            if config:
                payload["config"] = config
            
            logger.info(f"Creating integration for chat_id {chat_id} with app_id {app_id}")
            
            async with httpx.AsyncClient(timeout=30.0) as client:
                response = await client.post(
                    url,
                    json=payload,
                    headers={"Content-Type": "application/json"}
                )
                
                if response.status_code == 201:
                    integration_data = response.json()
                    logger.info(f"Successfully created integration {integration_data.get('integration_id')} for chat_id {chat_id}")
                    return integration_data
                else:
                    error_data = response.json() if response.headers.get('content-type', '').startswith('application/json') else {"error": response.text}
                    logger.error(f"API error creating integration for chat_id {chat_id}: {response.status_code} - {error_data}")
                    
                    if response.status_code == 400:
                        raise Exception(f"Invalid input: {error_data.get('message', 'Please check your data')}")
                    elif response.status_code == 404:
                        raise Exception("User or app not found")
                    elif response.status_code == 409:
                        raise Exception("Integration already exists for this app")
                    else:
                        raise Exception(f"Integration creation failed: {error_data.get('message', 'Unknown error')}")
                        
        except httpx.TimeoutException:
            logger.error(f"Timeout creating integration for chat_id {chat_id}")
            raise Exception("Server timeout - please try again later")
        except httpx.RequestError as e:
            logger.error(f"Request error creating integration for chat_id {chat_id}: {str(e)}")
            raise Exception("Unable to connect to server - please try again later")
        except Exception as e:
            if any(msg in str(e) for msg in ["Invalid input", "User or app not found", "Integration already exists", "Integration creation failed", "Server timeout", "Unable to connect"]):
                raise  # Re-raise our custom exceptions
            logger.error(f"Unexpected error creating integration for chat_id {chat_id}: {str(e)}")
            raise Exception("An unexpected error occurred - please try again later")

    async def test_integration(self, chat_id: str, integration_id: str, test_data: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Test an integration by creating a sample note.
        
        Args:
            chat_id: Telegram chat ID as string
            integration_id: Integration ID to test
            test_data: Optional test data
            
        Returns:
            Test result data
            
        Raises:
            Exception: If integration test fails
        """
        try:
            # Create a test note to verify the integration works
            test_note_data = {
                "title": "🧪 Test Note from Mainote Bot",
                "content": "This is a test note created during integration setup. If you see this, your integration is working correctly!",
                "category": "idea",
                "status": "active",
                "source": "telegram_integration_test"
            }
            
            if test_data:
                test_note_data.update(test_data)
            
            # Create the test note
            note_result = await self.create_note(
                chat_id=chat_id,
                title=test_note_data["title"],
                content=test_note_data["content"],
                category=test_note_data["category"],
                status=test_note_data["status"],
                source=test_note_data["source"],
                metadata={"integration_test": True, "integration_id": integration_id}
            )
            
            logger.info(f"Successfully tested integration {integration_id} for chat_id {chat_id}")
            
            return {
                "success": True,
                "message": "Integration test successful",
                "test_note": note_result
            }
            
        except Exception as e:
            logger.error(f"Integration test failed for integration {integration_id}: {str(e)}")
            return {
                "success": False,
                "message": f"Integration test failed: {str(e)}",
                "error": str(e)
            }
