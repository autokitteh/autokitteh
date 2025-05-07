import json
from io import StringIO
from datetime import datetime

import log


def test_log(monkeypatch):
    runner_id = "r1"

    io = StringIO()
    monkeypatch.setattr(log._handler, "stream", io)
    log.info("Grumpy")
    log.setup(runner_id)
    log.info("Garfield")

    io.seek(0)

    line = io.readline()
    rec = json.loads(line)
    assert rec["runner_id"] == "UNKNOWN RUNNER ID"
    assert rec["message"] == "Grumpy"

    line = io.readline()
    rec = json.loads(line)
    assert rec["runner_id"] == runner_id
    assert rec["message"] == "Garfield"

    # Check that time is in ISO format
    datetime.fromisoformat(rec["time"])
