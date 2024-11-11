import json

import hlog


def on_event(event):
    hlog.info(f"EVENT: {event}")
    data = json.loads(event.data.body.bytes)
    hlog.info(f"DATA: {data}")
