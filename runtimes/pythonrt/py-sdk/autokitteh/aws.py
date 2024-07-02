"""Initialize a Boto3 (AWS SDK) client, based on an AutoKitteh connection."""

import os

import boto3

from .connections import check_connection_name


def boto3_client(connection: str, service: str, region: str = "", **kwargs):
    """Initialize a Boto3 (AWS SDK) client, based on an AutoKitteh connection.

    API reference:
    https://boto3.amazonaws.com/v1/documentation/api/latest/index.html

    Code samples:
    https://boto3.amazonaws.com/v1/documentation/api/latest/guide/examples.html

    Args:
        connection: AutoKitteh connection name.
        service: AWS service name.
        region: AWS region name.

    Returns:
        Boto3 client.

    Raises:
        ValueError: AutoKitteh connection or AWS service/region names are invalid.
        BotoCoreError: Authentication error.
    """
    check_connection_name(connection)

    if not service:
        raise ValueError("AWS service name is required")

    if not region:
        region = os.getenv(connection + "__Region", "")

    return boto3.client(
        service,
        region,
        aws_access_key_id=os.getenv(connection + "__AccessKeyID"),
        aws_secret_access_key=os.getenv(connection + "__SecretKey"),
        aws_session_token=os.getenv(connection + "__Token"),
        **kwargs,
    )
