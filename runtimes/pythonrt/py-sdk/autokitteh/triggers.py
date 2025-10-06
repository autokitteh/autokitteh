"""Utility functions for triggers."""

import os


def get_webhook_url(trigger_name: str) -> str:
    url = os.getenv(f"{trigger_name}__webhook_url")
    if not url:
        raise ValueError(f"Webhook URL not found for trigger {trigger_name!r}")
    return url
