import os
import tempfile
from types import ModuleType

from kubernetes import client, config
from .connections import check_connection_name
from .errors import ConnectionInitError


def kubernetes_client(connection: str) -> ModuleType:
    """Initialize a Kubernetes client, based on an AutoKitteh connection.

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
        with tempfile.NamedTemporaryFile(delete=False, mode="w") as temp_file:
            temp_file.write(config_file)
            temp_path = temp_file.name

        config.load_kube_config(config_file=temp_path)
    except config.ConfigException as e:
        raise ConnectionInitError(connection) from e
    except Exception as e:
        raise ConnectionInitError(connection) from e

    finally:
        os.unlink(temp_path)

    return client
