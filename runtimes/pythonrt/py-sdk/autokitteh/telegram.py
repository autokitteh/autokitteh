"""Telegram bot client initialization and helper functions."""

import os
import re
from typing import Optional

import requests

from .connections import check_connection_name
from .errors import ConnectionInitError


class TelegramClient:
    """Simple Telegram Bot API client.
    
    This is a lightweight wrapper around the Telegram Bot API.
    For more advanced features, consider using libraries like python-telegram-bot.
    
    API reference:
    https://core.telegram.org/bots/api
    """
    
    def __init__(self, bot_token: str):
        """Initialize Telegram client with bot token.
        
        Args:
            bot_token: Telegram bot token from @BotFather.
        """
        self.bot_token = bot_token
        self.base_url = f"https://api.telegram.org/bot{bot_token}"
        self.session = requests.Session()
    
    def get_me(self) -> dict:
        """Get basic information about the bot.
        
        Returns:
            Bot information as dict.
            
        Raises:
            requests.RequestException: API request failed.
        """
        response = self.session.get(f"{self.base_url}/getMe")
        response.raise_for_status()
        return response.json()
    
    def send_message(self, chat_id: int, text: str, **kwargs) -> dict:
        """Send a text message.
        
        Args:
            chat_id: Unique identifier for the target chat.
            text: Text of the message to be sent.
            **kwargs: Additional parameters like parse_mode, reply_markup, etc.
            
        Returns:
            Sent message information as dict.
            
        Raises:
            requests.RequestException: API request failed.
        """
        data = {"chat_id": chat_id, "text": text, **kwargs}
        response = self.session.post(f"{self.base_url}/sendMessage", json=data)
        response.raise_for_status()
        return response.json()
    
    def send_photo(self, chat_id: int, photo: str, caption: Optional[str] = None, **kwargs) -> dict:
        """Send a photo.
        
        Args:
            chat_id: Unique identifier for the target chat.
            photo: Photo to send (file_id, URL, or file path).
            caption: Photo caption.
            **kwargs: Additional parameters.
            
        Returns:
            Sent message information as dict.
            
        Raises:
            requests.RequestException: API request failed.
        """
        data = {"chat_id": chat_id, "photo": photo, **kwargs}
        if caption:
            data["caption"] = caption
        response = self.session.post(f"{self.base_url}/sendPhoto", json=data)
        response.raise_for_status()
        return response.json()
    
    def send_document(self, chat_id: int, document: str, caption: Optional[str] = None, **kwargs) -> dict:
        """Send a document.
        
        Args:
            chat_id: Unique identifier for the target chat.
            document: Document to send (file_id, URL, or file path).
            caption: Document caption.
            **kwargs: Additional parameters.
            
        Returns:
            Sent message information as dict.
            
        Raises:
            requests.RequestException: API request failed.
        """
        data = {"chat_id": chat_id, "document": document, **kwargs}
        if caption:
            data["caption"] = caption
        response = self.session.post(f"{self.base_url}/sendDocument", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_updates(self, offset: Optional[int] = None, limit: Optional[int] = None, 
                   timeout: Optional[int] = None) -> dict:
        """Get incoming updates.
        
        Args:
            offset: Identifier of the first update to be returned.
            limit: Limits the number of updates to be retrieved.
            timeout: Timeout in seconds for long polling.
            
        Returns:
            Array of Update objects.
            
        Raises:
            requests.RequestException: API request failed.
        """
        params = {}
        if offset is not None:
            params["offset"] = offset
        if limit is not None:
            params["limit"] = limit
        if timeout is not None:
            params["timeout"] = timeout
            
        response = self.session.get(f"{self.base_url}/getUpdates", params=params)
        response.raise_for_status()
        return response.json()
    
    def set_webhook(self, url: str, **kwargs) -> dict:
        """Set a webhook URL to receive updates.
        
        Args:
            url: HTTPS URL to send updates to.
            **kwargs: Additional parameters like max_connections, allowed_updates, etc.
            
        Returns:
            Result of the operation.
            
        Raises:
            requests.RequestException: API request failed.
        """
        data = {"url": url, **kwargs}
        response = self.session.post(f"{self.base_url}/setWebhook", json=data)
        response.raise_for_status()
        return response.json()
    
    def delete_webhook(self) -> dict:
        """Delete the webhook integration.
        
        Returns:
            Result of the operation.
            
        Raises:
            requests.RequestException: API request failed.
        """
        response = self.session.post(f"{self.base_url}/deleteWebhook")
        response.raise_for_status()
        return response.json()
    
    def get_webhook_info(self) -> dict:
        """Get current webhook status.
        
        Returns:
            WebhookInfo object.
            
        Raises:
            requests.RequestException: API request failed.
        """
        response = self.session.get(f"{self.base_url}/getWebhookInfo")
        response.raise_for_status()
        return response.json()


def telegram_client(connection: str) -> TelegramClient:
    """Initialize a Telegram client, based on an AutoKitteh connection.

    API reference:
    https://core.telegram.org/bots/api

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Telegram bot client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)

    bot_token = os.getenv(connection + "__bot_token")
    if not bot_token:
        raise ConnectionInitError(connection)

    return TelegramClient(bot_token)


def bot_token(connection: str) -> str:
    """Get the bot token for a Telegram connection.
    
    Args:
        connection: AutoKitteh connection name.
        
    Returns:
        Bot token string.
        
    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)

    token = os.getenv(connection + "__bot_token")
    if not token:
        raise ConnectionInitError(connection)

    return token


def format_message(text: str, parse_mode: str = "Markdown") -> dict:
    """Format a message for Telegram with proper escaping.
    
    Args:
        text: Message text to format.
        parse_mode: Parse mode (Markdown, MarkdownV2, or HTML).
        
    Returns:
        Dict with text and parse_mode ready for sending.
    """
    if parse_mode == "Markdown":
        # Escape special characters for Markdown
        text = re.sub(r'([_*`\[])', r'\\\1', text)
    elif parse_mode == "HTML":
        # Escape HTML special characters
        text = text.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;')
    
    return {"text": text, "parse_mode": parse_mode}


def create_inline_keyboard(buttons: list) -> dict:
    """Create an inline keyboard markup.
    
    Args:
        buttons: List of button rows, where each row is a list of buttons.
                Each button is a dict with 'text' and either 'callback_data' or 'url'.
                
    Example:
        buttons = [
            [{"text": "Button 1", "callback_data": "data1"}],
            [{"text": "Button 2", "url": "https://example.com"}]
        ]
    
    Returns:
        Inline keyboard markup ready for use in send_message.
    """
    return {"inline_keyboard": buttons}


def create_reply_keyboard(buttons: list, **kwargs) -> dict:
    """Create a custom reply keyboard markup.
    
    Args:
        buttons: List of button rows, where each row is a list of button texts.
        **kwargs: Additional parameters like resize_keyboard, one_time_keyboard, etc.
                
    Example:
        buttons = [
            ["Yes", "No"],
            ["Maybe", "Cancel"]
        ]
    
    Returns:
        Reply keyboard markup ready for use in send_message.
    """
    keyboard = [[{"text": text} for text in row] for row in buttons]
    return {"keyboard": keyboard, **kwargs}
