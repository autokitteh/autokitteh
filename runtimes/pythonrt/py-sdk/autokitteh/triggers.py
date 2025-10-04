import os


def get_webhook_url(trigger_name: str) -> str:
    url = os.getenv(f"{trigger_name}__webhook_url")
    if not url:
        raise ValueError(f"No such webhook URL for trigger {trigger_name!r}, or trigger does not exist")
    return url
