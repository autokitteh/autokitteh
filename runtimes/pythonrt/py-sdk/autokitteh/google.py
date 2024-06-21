"""Initialize Google API clients, based on an AutoKitteh connections."""

from datetime import UTC, datetime
import json
import os
import re

# These imports work in AutoKitteh's virtual environment. The try-except block
# silences import errors in CI workflows that don't install these dependencies.
try:
    from google.auth.transport.requests import Request
    import google.oauth2.credentials as credentials
    import google.oauth2.service_account as service_account
    from googleapiclient.discovery import build
except ModuleNotFoundError:
    pass

from .connections import check_connection_name
from .errors import ConnectionInitError, EnvVarError


def gmail_client(connection: str, **kwargs):
    """Initialize a Gmail client, based on an AutoKitteh connection.

    API reference:
    https://developers.google.com/resources/api-libraries/documentation/gmail/v1/python/latest/gmail_v1.users.html

    Code samples:
    https://github.com/googleworkspace/python-samples/tree/main/gmail

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Gmail client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    # https://developers.google.com/gmail/api/auth/scopes
    default_scopes = [
        "https://www.googleapis.com/auth/gmail.modify",
        "https://www.googleapis.com/auth/gmail.settings.basic",
    ]
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

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    # https://developers.google.com/calendar/api/auth
    default_scopes = [
        "https://www.googleapis.com/auth/calendar",
        "https://www.googleapis.com/auth/calendar.events",
    ]
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

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    # https://developers.google.com/sheets/api/scopes
    default_scopes = ["https://www.googleapis.com/auth/spreadsheets"]
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

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    check_connection_name(connection)

    json_key = os.getenv(connection + "__JSON")  # Service Account (JSON key)
    if json_key:
        info = json.loads(json_key)
        # https://google-auth.readthedocs.io/en/stable/reference/google.oauth2.service_account.html#google.oauth2.service_account.Credentials.from_service_account_info
        return service_account.Credentials.from_service_account_info(
            info, scopes=scopes, **kwargs
        )

    refresh_token = os.getenv(connection + "__oauth_RefreshToken")  # User (OAuth 2.0)
    if refresh_token:
        return __google_creds_oauth2(connection, refresh_token, scopes)

    raise ConnectionInitError(connection)


def __google_creds_oauth2(connection: str, refresh_token: str, scopes: list[str]):
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

    Raises:
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Required environment variable is missing or invalid.
    """
    expiry = os.getenv(connection + "__oauth_Expiry")
    if not expiry:
        raise ConnectionInitError(connection)

    # Convert Go's time string (e.g. "2024-06-20 19:18:17 +0700 PDT") to
    # an ISO-8601 string that Python can parse with timezone awareness.
    timestamp = re.sub(r"[ A-Z]+.*", "", expiry)
    dt = datetime.fromisoformat(timestamp).astimezone(UTC)

    client_id = os.getenv("GOOGLE_CLIENT_ID")
    if not client_id:
        raise EnvVarError("GOOGLE_CLIENT_ID", "missing")

    client_secret = os.getenv("GOOGLE_CLIENT_SECRET")
    if not client_id:
        raise EnvVarError("GOOGLE_CLIENT_SECRET", "missing")

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
