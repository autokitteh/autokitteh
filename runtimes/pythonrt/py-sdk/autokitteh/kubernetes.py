import json
import os
from types import ModuleType

from kubernetes import client, config
from .connections import check_connection_name
from .errors import ConnectionInitError


def kubernetes_client(connection: str) -> ModuleType:
    """Initialize a Kubernetes client, based on an AutoKitteh connection.

    API reference:
    https://github.com/kubernetes-client/python
    https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Kubernetes API client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: If the connection config is missing or invalid,
            or if an unexpected error occurs during client initialization.

    """
    check_connection_name(connection)

    config_file = os.getenv(connection + "__configFile")

    if not config_file:
        raise ConnectionInitError(connection)

    try:
        config_dict = json.loads(config_file)
        config.load_kube_config_from_dict(config_dict)
    except config.ConfigException as e:
        raise ConnectionInitError(connection) from e
    except Exception as e:
        raise RuntimeError(
            f"Internal error while initializing Kubernetes client for connection '{connection}''"
        ) from e

    return client
