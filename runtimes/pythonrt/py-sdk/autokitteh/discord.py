import os

import discord

from .connections import check_connection_name
from .errors import ConnectionInitError


def discord_client(connection: str, intents=None, **kwargs) -> discord.Client:
    """Initialize a Discord client, based on an AutoKitteh connection.

    API reference:
    https://discordpy.readthedocs.io/en/stable/api.html

    Args:
        connection: AutoKitteh connection name.
        intents: An object representing the events your bot can receive.

    Returns:
        Discord client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        DiscordException: Connection attempt failed, or connection is unauthorized.
    """
    check_connection_name(connection)
    if not intents:
        intents = discord.Intents.default()

    bot_token = os.getenv(connection + "__BotToken")
    if not bot_token:
        raise ConnectionInitError(connection)

    return discord.Client(intents=intents, **kwargs)


def bot_token(connection: str):
    check_connection_name(connection)

    bot_token = os.getenv(connection + "__BotToken")
    if not bot_token:
        raise ConnectionInitError(connection)

    return bot_token
