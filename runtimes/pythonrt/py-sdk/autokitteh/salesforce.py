"""Initialize a Salesforce client, based on an AutoKitteh connection."""

import os

from .connections import check_connection_name
from .errors import ConnectionInitError, EnvVarError

from simple_salesforce import Salesforce


def salesforce_client(connection: str, **kwargs) -> Salesforce:
    """Initialize a Salesforce client, based on an AutoKitteh connection.

    API reference:
    https://github.com/simple-salesforce/simple-salesforce

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Salesforce client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        EnvVarError: Salesforce instance URL is missing.
        TypeError: Invalid auth credentials.
    """
    check_connection_name(connection)

    oauth_token = os.getenv(connection + "__oauth_AccessToken")
    if not oauth_token:
        raise ConnectionInitError(connection)

    # TODO: Get instance URL from connection vars.
    instance_url = os.getenv("SALESFORCE_INSTANCE_URL")
    if not instance_url:
        raise EnvVarError("SALESFORCE_INSTANCE_URL", "missing")

    return Salesforce(instance_url=instance_url, oauth_token=oauth_token, **kwargs)
