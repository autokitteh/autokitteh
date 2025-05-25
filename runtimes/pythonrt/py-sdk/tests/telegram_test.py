"""Tests for Telegram client functionality."""

import os
import pytest
from unittest.mock import Mock, patch

from autokitteh.telegram import telegram_client, bot_token, TelegramClient, format_message, create_inline_keyboard
from autokitteh.errors import ConnectionInitError


class TestTelegramClient:
    """Test the TelegramClient class."""
    
    def test_init(self):
        """Test client initialization."""
        client = TelegramClient("123456:ABC-DEF")
        assert client.bot_token == "123456:ABC-DEF"
        assert client.base_url == "https://api.telegram.org/bot123456:ABC-DEF"
    
    @patch('requests.Session.get')
    def test_get_me(self, mock_get):
        """Test getMe API call."""
        mock_response = Mock()
        mock_response.json.return_value = {
            "ok": True,
            "result": {
                "id": 123456789,
                "is_bot": True,
                "first_name": "Test Bot",
                "username": "testbot"
            }
        }
        mock_get.return_value = mock_response
        
        client = TelegramClient("test-token")
        result = client.get_me()
        
        assert result["ok"] is True
        assert result["result"]["id"] == 123456789
        assert result["result"]["username"] == "testbot"
        mock_get.assert_called_once_with("https://api.telegram.org/bottest-token/getMe")
    
    @patch('requests.Session.post')
    def test_send_message(self, mock_post):
        """Test sendMessage API call."""
        mock_response = Mock()
        mock_response.json.return_value = {
            "ok": True,
            "result": {
                "message_id": 1,
                "date": 1609459200,
                "chat": {"id": 123, "type": "private"},
                "text": "Hello, World!"
            }
        }
        mock_post.return_value = mock_response
        
        client = TelegramClient("test-token")
        result = client.send_message(123, "Hello, World!")
        
        assert result["ok"] is True
        assert result["result"]["text"] == "Hello, World!"
        mock_post.assert_called_once_with(
            "https://api.telegram.org/bottest-token/sendMessage",
            json={"chat_id": 123, "text": "Hello, World!"}
        )


class TestHelperFunctions:
    """Test helper functions."""
    
    @patch.dict(os.environ, {"test_conn__bot_token": "123456:ABC-DEF"})
    def test_telegram_client_success(self):
        """Test successful telegram_client creation."""
        client = telegram_client("test_conn")
        assert isinstance(client, TelegramClient)
        assert client.bot_token == "123456:ABC-DEF"
    
    def test_telegram_client_missing_token(self):
        """Test telegram_client with missing token."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(ConnectionInitError):
                telegram_client("test_conn")
    
    @patch.dict(os.environ, {"test_conn__bot_token": "123456:ABC-DEF"})
    def test_bot_token_success(self):
        """Test successful bot_token retrieval."""
        token = bot_token("test_conn")
        assert token == "123456:ABC-DEF"
    
    def test_bot_token_missing(self):
        """Test bot_token with missing token."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(ConnectionInitError):
                bot_token("test_conn")
    
    def test_format_message_markdown(self):
        """Test message formatting for Markdown."""
        result = format_message("Hello *world*!", "Markdown")
        assert result["text"] == "Hello \\*world\\*!"
        assert result["parse_mode"] == "Markdown"
    
    def test_format_message_html(self):
        """Test message formatting for HTML."""
        result = format_message("Hello <b>world</b>!", "HTML")
        assert result["text"] == "Hello &lt;b&gt;world&lt;/b&gt;!"
        assert result["parse_mode"] == "HTML"
    
    def test_create_inline_keyboard(self):
        """Test inline keyboard creation."""
        buttons = [
            [{"text": "Button 1", "callback_data": "data1"}],
            [{"text": "Button 2", "url": "https://example.com"}]
        ]
        result = create_inline_keyboard(buttons)
        
        expected = {"inline_keyboard": buttons}
        assert result == expected
