"""AutoKitteh Python SDK."""

from datetime import UTC, datetime
import json
import os
import re

try:
    from atlassian import Jira
    from google.auth.transport.requests import Request
    import google.oauth2.credentials as credentials
    import google.oauth2.service_account as service_account
    from googleapiclient.discovery import build
    import slack_sdk
except ModuleNotFoundError:
    pass  # These imports will work in AutoKitteh's virtual environment.

from uuid import uuid4


ACTIVITY_ATTR = "__activity__"

class AttrDict(dict):
    """Allow attribute access to dictionary keys.

    >>> config = AttrDict({'server': {'port': 8080}, 'debug': True})
    >>> config.server.port
    8080
    >>> config.debug
    True
    """
    def __getattr__(self, name):
        try:
            value = self[name]
            if isinstance(value, dict):
                value = AttrDict(value)
            return value
        except KeyError:
            raise AttributeError(name)

    def __setattr__(self, attr, value):
        # The default __getattr__ doesn't fail but also don't change values
        cls = self.__class__.__name__
        raise NotImplementedError(f'{cls} does not support setting attributes')


def activity(fn: callable) -> callable:
    """Decorator to mark a function as an activity."""
    setattr(fn, ACTIVITY_ATTR, True)
    return fn


<<<<<<< HEAD
def subscribe(connection_name: str, filter: str) -> str:
    """Subscribe to events on connection. Option filter is a CEL expression."""
=======
def jira_client(connection: str, **kwargs) -> Jira:
    """Initialize a Jira client, based on an AutoKitteh connection.

    API reference:
    https://atlassian-python-api.readthedocs.io/jira.html

    Code samples:
    https://github.com/atlassian-api/atlassian-python-api/tree/master/examples/jira

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Atlassian-Python-API Jira client.
    """
    if not re.fullmatch(r"[A-Za-z_]\w*", connection):
        raise ValueError(f'Invalid AutoKitteh connection name: "{connection}"')

    if os.getenv(connection + "__oauth_AccessToken"):
        return _jira_client_cloud_oauth2(connection, **kwargs)

    raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')


def _jira_client_cloud_oauth2(connection: str, **kwargs) -> Jira:
    """Initialize a Jira client for Atlassian Cloud using OAuth 2.0."""
    expiry = os.getenv(connection + "__oauth_Expiry")
    if not expiry:
        raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')

    iso8601 = re.sub(r"[ A-Z]+$", "", expiry)  # Convert from Go's time string.
    if datetime.fromisoformat(iso8601) < datetime.now(UTC):
        raise RuntimeError("OAuth 2.0 access token expired on: " + expiry)

    cloud_id = os.getenv(connection + "__access_id")
    if not cloud_id:
        raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')

    client_id = os.getenv("JIRA_CLIENT_ID")
    if not client_id:
        raise RuntimeError('Environment variable "JIRA_CLIENT_ID" not set')

    return Jira(
        url="https://api.atlassian.com/ex/jira/" + cloud_id,
        oauth2={
            "client_id": client_id,
            "token": {
                "access_token": os.getenv(connection + "__oauth_AccessToken"),
                "token_type": os.getenv(connection + "__oauth_TokenType"),
            },
        },
        **kwargs,
    )


def gmail_client(connection: str, **kwargs):
    """Initialize a Gmail client, based on an AutoKitteh connection.

    API reference:
    https://developers.google.com/resources/api-libraries/documentation/gmail/v1/python/latest/gmail_v1.users.html

    Code samples:
    https://github.com/googleworkspace/python-samples/tree/main/gmail

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Google Sheets client.
    """
    default_scopes = []
    creds = google_creds(connection, default_scopes, **kwargs)
    return build("gmail", "v1", credentials=creds, **kwargs)


def google_calendar_client(connection: str, **kwargs):
    """Initialize a Google Calendar client, based on an AutoKitteh connection.

    API reference:
    https://developers.google.com/resources/api-libraries/documentation/calendar/v3/python/latest/index.html

    Code samples:
    https://github.com/googleworkspace/python-samples/tree/main/calendar

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Google Calendar client.
    """
    default_scopes = []
    creds = google_creds(connection, default_scopes, **kwargs)
    return build("calendar", "v3", credentials=creds, **kwargs)


def google_sheets_client(connection: str, **kwargs):
    """Initialize a Google Sheets client, based on an AutoKitteh connection.

    API reference:
    https://developers.google.com/resources/api-libraries/documentation/sheets/v4/python/latest/sheets_v4.spreadsheets.html

    Code samples:
    https://github.com/googleworkspace/python-samples/tree/main/sheets

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Google Sheets client.
    """
    default_scopes = []
    creds = google_creds(connection, default_scopes, **kwargs)
    return build("sheets", "v4", credentials=creds, **kwargs)


def google_creds(connection: str, scopes: list[str], **kwargs):
    """Initialize credentials for a Google APIs client, for service discovery.

    This function supports both AutoKitteh connection modes:
    users (with OAuth 2.0), and GCP service accounts (with a JSON key).

    Code samples:
    https://github.com/googleworkspace/python-samples

    For subsequent usage details, see:
    https://googleapis.github.io/google-api-python-client/docs/epy/googleapiclient.discovery-module.html#build

    Args:
        connection: AutoKitteh connection name.
        scopes: List of OAuth permission scopes.

    Returns:
        Google API credentials, ready for usage
        in "googleapiclient.discovery.build()".
    """
    if not re.fullmatch(r"[A-Za-z_]\w*", connection):
        raise ValueError(f'Invalid AutoKitteh connection name: "{connection}"')

    json_key = os.getenv(connection + "__JSON")  # Service Account (JSON key)
    if json_key:
        info = json.loads(json_key)
        # https://google-auth.readthedocs.io/en/stable/reference/google.oauth2.service_account.html#google.oauth2.service_account.Credentials.from_service_account_info
        return service_account.Credentials.from_service_account_info(
            info, scopes=scopes, **kwargs
        )

    refresh_token = os.getenv(connection + "__oauth_RefreshToken")  # User (OAuth 2.0)
    if refresh_token:
        return _google_creds_oauth2(connection, refresh_token, scopes)

    raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')


def _google_creds_oauth2(connection: str, refresh_token: str, scopes: list[str]):
    """Initialize user credentials for Google APIs using OAuth 2.0.

    For more details, see:
    https://google-auth.readthedocs.io/en/stable/reference/google.oauth2.credentials.html#google.oauth2.credentials.Credentials.from_authorized_user_info

    Args:
        connection: AutoKitteh connection name.
        refresh_token: OAuth 2.0 refresh token.
        scopes: List of OAuth permission scopes.

    Returns:
        Google API credentials, ready for usage
        in "googleapiclient.discovery.build()".
    """
    expiry = os.getenv(connection + "__oauth_Expiry")
    iso8601 = re.sub(r"[ A-Z]+$", "", expiry)  # Convert from Go's time string.
    dt = datetime.fromisoformat(iso8601).astimezone(UTC)

    client_id = os.getenv("GOOGLE_CLIENT_ID")
    if not client_id:
        raise RuntimeError('Environment variable "GOOGLE_CLIENT_ID" not set')

    client_secret = os.getenv("GOOGLE_CLIENT_SECRET")
    if not client_id:
        raise RuntimeError('Environment variable "GOOGLE_CLIENT_SECRET" not set')

    creds = credentials.Credentials.from_authorized_user_info(
        {
            "token": os.getenv(connection + "__oauth_AccessToken"),
            "refresh_token": refresh_token,
            "expiry": dt.replace(tzinfo=None).isoformat(),
            "client_id": client_id,
            "client_secret": client_secret,
            "scopes": scopes,
        }
    )
    if creds.expired:
        creds.refresh(Request())

    return creds


def slack_client(connection: str, **kwargs) -> slack_sdk.web.client.WebClient:
    """Initialize a Slack client, based on an AutoKitteh connection.
>>>>>>> cce9af3b (Jira with OAuth 2.0)

    # Dummy implementation for local development
    return uuid4().hex


def unsubscribe(id: str) -> None:
    """Unsubscribe from events."""

<<<<<<< HEAD
    # Dummy implementation for local development
    pass
=======
    Returns:
        Slack SDK client.
    """
    if not re.fullmatch(r"[A-Za-z_]\w*", connection):
        raise ValueError(f'Invalid AutoKitteh connection name: "{connection}"')

    bot_token = os.getenv(connection + "__oauth_AccessToken")  # OAuth v2
    if not bot_token:
        bot_token = os.getenv(connection + "__BotToken")  # Socket Mode
    if not bot_token:
        raise RuntimeError(f'AutoKitteh connection "{connection}" not initialized')
>>>>>>> cce9af3b (Jira with OAuth 2.0)


def next_event(id: str) -> AttrDict:
    """Get the next event from the subscription."""

    # Dummy implementation for local development
    return AttrDict()
