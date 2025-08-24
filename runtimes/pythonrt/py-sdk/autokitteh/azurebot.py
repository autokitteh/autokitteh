"""Initialize a Twilio client, based on an AutoKitteh connection."""

from dataclasses import dataclass
from typing import Any
import os
import requests

import O365

from .connections import check_connection_name
from .errors import ConnectionInitError, AuthenticationError


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


def azurebot_o365_account(connection: str) -> O365.Account:
    """Initialize an Azure Bot client, based on an AutoKitteh connection.

    API reference:
        https://github.com/O365/python-o365

    Args:
        connection: AutoKitteh connection name.

    Returns:
        O365 Account object.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
    """
    check_connection_name(connection)
    
    credentials = _get_credentials(connection)

    account = O365.Account(
        (credentials.app_id, credentials.app_password), 
        auth_flow_type='credentials', 
        tenant_id=credentials.tenant_id,
    )

    if not account.authenticate():
        raise AuthenticationError(connection)

    return account


class AzureBotClient:
    def __init__(self, connection: str):
        self._credentials = _get_credentials(connection)

    def _get_access_token(self) -> str:
        url = f"https://login.microsoftonline.com/{self._credentials.tenant_id}/oauth2/v2.0/token"
        data = {
            'grant_type': 'client_credentials',
            'client_id': self._credentials.app_id,
            'client_secret': self._credentials.app_password,
            'scope': 'https://api.botframework.com/.default'
        }
        
        response = requests.post(url, data=data)
        if response.status_code != 200:
            # TODO: More specific exception.
            raise Exception(f"Failed to get token: {response.text}")
        
        token_data = response.json()
        if token_data.get('error'):
            raise Exception(f"Failed to get token: {token_data['error']}")
        
        return token_data['access_token']

    def send_activity(self, service_url, conversation_id, activity) -> Any:
        """Send activity synchronously"""
        token = self._get_access_token()
        url = f"{service_url}v3/conversations/{conversation_id}/activities"
        
        print(token)
        
        headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
        
        response = requests.post(url, json=activity, headers=headers)
        response.raise_for_status()

        return response.json()


def azurebot_client(connection: str) -> AzureBotClient:
    return AzureBotClient(connection)
