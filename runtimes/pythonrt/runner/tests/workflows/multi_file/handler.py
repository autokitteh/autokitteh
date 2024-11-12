import json
from uuid import uuid4

import hlog


def on_event(event):
    hlog.info(f"EVENT: {event}")
    id = str(uuid4().hex)
    print(json.dumps({"id": id}))
