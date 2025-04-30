import json
import logging
from logging import LogRecord
from os import environ
import sys

import pb


class AKHandler(logging.Handler):
    def __init__(self):
        super().__init__()
        self.runner_id = ""
        self.worker: pb.handler_rpc.HandlerServiceStub = None

    def setup(self, runner_id, worker):
        self.runner_id = runner_id
        self.worker = worker

    def emit(self, record: LogRecord) -> None:
        message = self.format(record)
        if not self.worker:
            print(f"ERROR: log without worker: {message}", file=sys.stderr)
            return

        req = pb.handler.LogRequest(
            runner_id=self.runner_id,
            level=record.levelname,
            message=message,
        )

        resp = self.worker.Log(req)
        if resp.error != "":
            print(f"ERROR: log error: {resp.error}", file=sys.stderr)


class JSONFormatter(logging.Formatter):
    """Formatter to dump error message into JSON"""

    def format(self, record: LogRecord) -> str:
        record_dict = {
            "level": record.levelname,
            "date": self.formatTime(record),
            "message": record.getMessage(),
            "filename": record.filename,
            "lineno": record.lineno,
            "function": record.funcName,
        }
        return json.dumps(record_dict)


_log = logging.getLogger("ak")
_log.setLevel(environ.get("AK_WORKER_PYTHON_LOG_LEVEL") or logging.INFO)
_handler = AKHandler()
_handler.setFormatter(JSONFormatter())
_log.addHandler(_handler)

info = _log.info
warning = _log.warning
error = _log.error
exception = _log.exception


def setup(runner_id, worker):
    _handler.setup(runner_id, worker)
