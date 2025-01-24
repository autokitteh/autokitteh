"""Initialize Google API clients, based on AutoKitteh connections."""

from datetime import UTC, datetime
import json
import os
import re

from google.auth.exceptions import RefreshError
from google.auth.transport.requests import Request
import google.generativeai as genai
import google.oauth2.credentials as credentials
import google.oauth2.service_account as service_account
from googleapiclient.discovery import build

from .connections import check_connection_name, refresh_oauth
from .errors import ConnectionInitError, OAuthRefreshError


def gmail_client(connection: str, **kwargs):
    """Initialize a Gmail client, based on an AutoKitteh connection.

    API documentation:
    https://docs.autokitteh.com/integrations/google/gmail/python

    Code samples:
    - https://github.com/autokitteh/kittehub/tree/main/samples/google/gmail
    - https://github.com/googleworkspace/python-samples/tree/main/gmail

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Gmail client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    # https://developers.google.com/gmail/api/auth/scopes
    default_scopes = [
        "https://www.googleapis.com/auth/gmail.modify",
        "https://www.googleapis.com/auth/gmail.settings.basic",
    ]
    creds = google_creds("gmail", connection, default_scopes, **kwargs)
    return build("gmail", "v1", credentials=creds, **kwargs)


def google_calendar_client(connection: str, **kwargs):
    """Initialize a Google Calendar client, based on an AutoKitteh connection.

    API documentation:
    https://docs.autokitteh.com/integrations/google/calendar/python

    Code samples:
    https://github.com/autokitteh/kittehub/tree/main/samples/google/calendar

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Google Calendar client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    # https://developers.google.com/calendar/api/auth
    default_scopes = [
        "https://www.googleapis.com/auth/calendar",
        "https://www.googleapis.com/auth/calendar.events",
    ]
    creds = google_creds("googlecalendar", connection, default_scopes, **kwargs)
    return build("calendar", "v3", credentials=creds, **kwargs)


def google_drive_client(connection: str, **kwargs):
    """Initialize a Google Drive client, based on an AutoKitteh connection.

    API documentation:
    https://docs.autokitteh.com/integrations/google/drive/python

    Code samples:
    https://github.com/googleworkspace/python-samples/tree/main/drive

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Google Drive client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    # https://developers.google.com/drive/api/guides/api-specific-auth
    default_scopes = [
        "https://www.googleapis.com/auth/drive.file",  # See ENG-1701
    ]
    creds = google_creds("googledrive", connection, default_scopes, **kwargs)
    return build("drive", "v3", credentials=creds, **kwargs)


def google_forms_client(connection: str, **kwargs):
    """Initialize a Google Forms client, based on an AutoKitteh connection.

    API documentation:
    https://docs.autokitteh.com/integrations/google/forms/python

    Code samples:
    - https://github.com/autokitteh/kittehub/tree/main/samples/google/forms
    - https://github.com/googleworkspace/python-samples/tree/main/forms

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Google Forms client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    # https://developers.google.com/identity/protocols/oauth2/scopes#script
    default_scopes = [
        "https://www.googleapis.com/auth/forms.body",
        "https://www.googleapis.com/auth/forms.responses.readonly",
    ]
    creds = google_creds("googleforms", connection, default_scopes, **kwargs)
    return build("forms", "v1", credentials=creds, **kwargs)


def gemini_client(connection: str, **kwargs) -> genai.GenerativeModel:
    """Initialize a Gemini generative AI client, based on an AutoKitteh connection.

    API reference:
    - https://ai.google.dev/gemini-api/docs
    - https://github.com/google-gemini/generative-ai-python/blob/main/docs/api/google/generativeai/GenerativeModel.md

    Code samples:
    - https://ai.google.dev/gemini-api/docs#explore-the-api
    - https://ai.google.dev/gemini-api/docs/text-generation?lang=python
    - https://github.com/google-gemini/generative-ai-python/tree/main/samples
    - https://github.com/google-gemini/cookbook

    Args:
        connection: AutoKitteh connection name.

    Returns:
        An initialized GenerativeModel instance.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)

    # Set the API key, if possible.
    api_key = os.getenv(connection + "__api_key")
    if not api_key:
        raise ConnectionInitError(connection)

    genai.configure(api_key=api_key)
    return genai.GenerativeModel(**kwargs)


def google_sheets_client(connection: str, **kwargs):
    """Initialize a Google Sheets client, based on an AutoKitteh connection.

    API documentation:
    https://docs.autokitteh.com/integrations/google/sheets/python

    Code samples:
    - https://github.com/autokitteh/kittehub/tree/main/samples/google/sheets
    - https://github.com/googleworkspace/python-samples/tree/main/sheets

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Google Sheets client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    # https://developers.google.com/sheets/api/scopes
    default_scopes = ["https://www.googleapis.com/auth/spreadsheets"]
    creds = google_creds("googlesheets", connection, default_scopes, **kwargs)
    return build("sheets", "v4", credentials=creds, **kwargs)


def google_creds(integration: str, connection: str, scopes: list[str], **kwargs):
    """Initialize credentials for a Google APIs client, for service discovery.

    This function supports both AutoKitteh connection modes:
    users (with OAuth 2.0), and GCP service accounts (with a JSON key).

    Code samples:
    https://github.com/googleworkspace/python-samples

    For subsequent usage details, see:
    https://googleapis.github.io/google-api-python-client/docs/epy/googleapiclient.discovery-module.html#build

    Args:
        integration: AutoKitteh integration name.
        connection: AutoKitteh connection name.
        scopes: List of OAuth permission scopes.

    Returns:
        Google API credentials, ready for usage
        in "googleapiclient.discovery.build()".

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    check_connection_name(connection)

    if os.getenv(connection + "__authType") == "oauth":  # User (OAuth 2.0)
        return _google_creds_oauth2(integration, connection, scopes)

    json_key = os.getenv(connection + "__JSON")  # Service Account (JSON key)
    if json_key:
        # https://google-auth.readthedocs.io/en/stable/reference/google.oauth2.service_account.html#google.oauth2.service_account.Credentials.from_service_account_info
        return service_account.Credentials.from_service_account_info(
            json.loads(json_key), scopes=scopes, **kwargs
        )

    raise ConnectionInitError(connection)


def _google_creds_oauth2(integration: str, connection: str, scopes: list[str]):
    """Initialize user credentials for Google APIs using OAuth 2.0.

    For more details, see:
    - https://google-auth.readthedocs.io/en/stable/reference/google.oauth2.credentials.html#google.oauth2.credentials.Credentials.from_authorized_user_info
    - https://github.com/googleapis/google-auth-library-python/blob/main/google/oauth2/credentials.py

    Args:
        integration: AutoKitteh integration name.
        connection: AutoKitteh connection name.
        scopes: List of OAuth permission scopes.

    Returns:
        Google API credentials, ready for usage
        in "googleapiclient.discovery.build()".

    Raises:
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        OAuthRefreshError: OAuth token refresh failed.
    """
    expiry = os.getenv(connection + "__oauth_Expiry")
    if not expiry:
        raise ConnectionInitError(connection)

    # Convert Go's time string (e.g. "2024-06-20 19:18:17 -0700 PDT") to
    # an ISO-8601 string that Python can parse with timezone awareness.
    expiry = re.sub(r" [A-Z]+.*", "", expiry)
    expiry = re.sub(r"\.\d+", "", expiry)  # Also ignore sub-second precision.
    dt = datetime.fromisoformat(expiry).astimezone(UTC).replace(tzinfo=None)

    token = os.getenv(connection + "__oauth_AccessToken")
    client_secret = os.getenv("GOOGLE_CLIENT_SECRET", "NOT AVAILABLE")

    if client_secret == "NOT AVAILABLE":
        # In Docker/Cloud environments, handle token refresh through AutoKitteh
        creds = credentials.Credentials(token=token, expiry=dt, scopes=scopes)
        creds.refresh_handler = _google_refresh_handler(integration, connection)
    else:
        # Refreshes to be handled by the Python client, as usual.
        creds = credentials.Credentials.from_authorized_user_info(
            {
                "token": token,
                "refresh_token": os.getenv(connection + "__oauth_RefreshToken"),
                "expiry": dt.isoformat(),
                "client_id": os.getenv("GOOGLE_CLIENT_ID"),
                "client_secret": client_secret,
                "scopes": scopes,
            }
        )

    try:
        if creds.expired:
            creds.refresh(Request())
    except RefreshError as e:
        raise OAuthRefreshError(connection, e)

    return creds


def _google_refresh_handler(integration: str, connection: str) -> callable:
    """Refresh handler for OAuth 2.0 user credentials for Google APIs.

    For more details, see:
    - https://google-auth.readthedocs.io/en/stable/reference/google.oauth2.credentials.html
    - https://github.com/googleapis/google-auth-library-python/blob/main/google/oauth2/credentials.py

    Args:
        integration: AutoKitteh integration name.
        connection: AutoKitteh connection name.

    Returns:
        Generated function (based on the input) to return a fresh access token
        and its expiry date. Overridden by AutoKitteh to keep the Google client
        secret hidden from workflows, and unused in local development.
    """

    def impl(request, scopes: list[str]) -> tuple[str, datetime]:
        return refresh_oauth(integration, connection)

    return impl


def google_id(url: str) -> str:
    """Extract the Google Doc/Form/Sheet ID from a URL. This function is idempotent.

    Example: 'https://docs.google.com/.../d/1a2b3c4d5e6f/edit' --> '1a2b3c4d5e6f'
    """
    match = re.match(r"(.*/d/(e/)?)?([\w-]{20,})", url)
    if match:
        return match.group(3)
    else:
        raise ValueError(f'Invalid Google ID in "{url}"')
