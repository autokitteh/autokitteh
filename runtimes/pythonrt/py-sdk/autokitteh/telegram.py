"""Initialize a Telegram client, based on an AutoKitteh connection."""

import os

from telegram import Bot


from .connections import check_connection_name
from .errors import ConnectionInitError


def telegram_client(connection: str) -> Bot:
    """Initialize a Telegram client, based on an AutoKitteh connection.

    API reference:
        https://github.com/python-telegram-bot/python-telegram-bot
        https://core.telegram.org/bots/api
        https://core.telegram.org/bots/samples

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Telegram Bot API client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        telegram.error.TelegramError: Telegram SDK initialization errors.
    """
    check_connection_name(connection)

    bot_token = os.getenv(connection + "__BotToken")

    if not bot_token:
        raise ConnectionInitError(connection)

    return Bot(token=bot_token)
