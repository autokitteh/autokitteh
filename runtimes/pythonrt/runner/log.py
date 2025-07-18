import json
import logging
from logging import LogRecord
from os import environ


class JsonFormatter(logging.Formatter):
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


formatter = JsonFormatter()

_log = logging.getLogger()
_log.setLevel(environ.get("AK_WORKER_PYTHON_LOG_LEVEL") or logging.INFO)
_stream_handler = logging.StreamHandler()
_stream_handler.setFormatter(formatter)
_log.addHandler(_stream_handler)

info = _log.info
debug = _log.debug
warning = _log.warning
error = _log.error
exception = _log.exception
