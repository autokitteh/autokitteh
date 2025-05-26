import json
import os
import tempfile

from kubernetes import client, config
from .connections import check_connection_name
from .errors import ConnectionInitError


def kubernetes_client(connection: str):  # TODO: add type hints
    """Initialize a Kubernetes client, based on an AutoKitteh connection.

    Args:
        connection: AutoKitteh connection name.

    Returns:
        Kubernetes API client.

    Raises:
        ValueError: AutoKitteh connection name is invalid.
        ConnectionInitError: AutoKitteh connection was not initialized yet.

    """
    check_connection_name(connection)

    config_file = os.getenv(connection + "__configFile")

    if not config_file:
        raise ConnectionInitError(connection)

    parsed = json.loads(config_file)
    with tempfile.NamedTemporaryFile(delete=False, mode="w") as temp_file:
        json.dump(parsed, temp_file, indent=2)
        temp_path = temp_file.name

    try:
        config.load_kube_config(config_file=temp_path)
    except config.ConfigException as e:
        raise ConnectionInitError(connection) from e
    finally:
        os.unlink(temp_path)

    return client
