"""Initialize a AzureBot client, based on an AutoKitteh connection."""

from dataclasses import dataclass
from typing import Any
import os
import requests

from .connections import check_connection_name
from .errors import ConnectionInitError, AuthenticationError
from .activities import activity


@dataclass
class _Credentials:
    app_id: str
    app_password: str
    tenant_id: str


def _get_credentials(connection: str) -> _Credentials:
    app_id = os.getenv(connection + "__app_id")
    app_password = os.getenv(connection + "__app_password")
    tenant_id = os.getenv(connection + "__tenant_id")
    if not app_id or not app_password or not tenant_id:
        raise ConnectionInitError(connection)
    return _Credentials(app_id, app_password, tenant_id)


class AzureBotClient:
    def __init__(self, connection: str):
        self._connection = connection
        self._credentials = _get_credentials(connection)

    def _get_access_token(self) -> str:
        url = f"https://login.microsoftonline.com/{self._credentials.tenant_id}/oauth2/v2.0/token"
        data = {
            "grant_type": "client_credentials",
            "client_id": self._credentials.app_id,
            "client_secret": self._credentials.app_password,
            "scope": "https://api.botframework.com/.default",
        }

        response = requests.post(url, data=data)
        if response.status_code != 200:
            raise AuthenticationError(
                self._connection, f"Failed to get token: {response.text}"
            )

        token_data = response.json()
        if token_data.get("error"):
            raise AuthenticationError(
                self._connection, f"Failed to get token: {token_data['error']}"
            )

        return token_data["access_token"]

    @activity
    def send_conversation_activity(
        self,
        activity: dict,
        conversation_id: str | None,
        service_url: str = "https://smba.trafficmanager.net/teams/",
    ) -> Any:
        """Send activity synchronously.

        If this is sent as a reply to an event, use the service_url from that event.

        Raises on non-2xx statuses.

        Returns the HTTP response body as JSON.
        """
        token = self._get_access_token()

        if conversation_id:
            url = f"{service_url}v3/conversations/{conversation_id}/activities"
        else:
            url = f"{service_url}v3/conversations"

        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
        }

        response = requests.post(url, json=activity, headers=headers)
        response.raise_for_status()
        return response.json()


def azurebot_client(connection: str) -> AzureBotClient:
    check_connection_name(connection)

    return AzureBotClient(connection)
