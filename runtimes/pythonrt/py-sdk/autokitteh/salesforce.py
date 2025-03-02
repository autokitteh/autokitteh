"""Initialize a Salesforce client, based on an AutoKitteh connection."""

import os

from simple_salesforce import Salesforce

from .connections import check_connection_name
from .errors import ConnectionInitError


def salesforce_client(connection: str, **kwargs) -> Salesforce:
    """Initialize a Salesforce client, based on an AutoKitteh connection.

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Salesforce client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.
        SalesforceApiError: Connection attempt failed, or connection is unauthorized.
    """
    check_connection_name(connection)

    token = os.getenv(connection + "__oauth_access_token")
    instance_url = os.getenv(connection + "__instance_url")
    if not token or not instance_url:
        raise ConnectionInitError(connection)

    return Salesforce(instance_url=instance_url, session_id=token, **kwargs)
