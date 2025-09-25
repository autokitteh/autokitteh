import json
import logging
import sys
from datetime import datetime
from logging import LogRecord
from os import environ


class JSONFormatter(logging.Formatter):
    """Formatter to dump error message into JSON"""

    runner_id = "UNKNOWN RUNNER ID"

    def format(self, record: LogRecord) -> str:
        record_dict = {
            "level": record.levelname,
            "time": datetime.fromtimestamp(record.created).isoformat(),
            "message": record.getMessage(),
            "filename": record.filename,
            "lineno": record.lineno,
            "function": record.funcName,
            "runner_id": self.runner_id,
            "app": record.name,
        }
        return json.dumps(record_dict)


_formatter = JSONFormatter()

_log = logging.getLogger("Python runner")
_log.setLevel(environ.get("AK_WORKER_PYTHON_LOG_LEVEL") or logging.INFO)
_handler = logging.StreamHandler(sys.stdout)
_handler.setFormatter(_formatter)
_log.addHandler(_handler)

info = _log.info
debug = _log.debug
warning = _log.warning
error = _log.error
exception = _log.exception


def setup(runner_id):
    _formatter.runner_id = runner_id
