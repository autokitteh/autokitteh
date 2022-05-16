from typing import Any

from collections import namedtuple

from twilio.rest import Client

from autokitteh.api import PluginID
import autokitteh.plugin


__all__ = ['Plugin']


TwilioClient = namedtuple(
    'TwilioClient',
    [
        'create_message',
    ],
)


def _open(account_sid: str, auth_token: str) -> TwilioClient:
    client = Client(account_sid, auth_token)

    def create_message(*args: Any, **kwargs: Any) -> Any:
        return client.messages.create(*args, **kwargs)._properties

    return TwilioClient(
        create_message=create_message,
    )


Plugin = autokitteh.plugin.Plugin(
    id=PluginID("twilio"),
    doc="Twilio plugin",
    members={
        "open": _open,
    },
)
