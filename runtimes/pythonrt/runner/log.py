import logging
from logging import LogRecord
import json


class JsonFormatter(logging.Formatter):
    """Formatter to dump error message into JSON"""

    def format(self, record: LogRecord) -> str:
        record_dict = {
            "level": record.levelname,
            "date": self.formatTime(record),
            "message": record.getMessage(),
            "filename": record.filename,
            "lineno": record.lineno,
        }
        return json.dumps(record_dict)


formatter = JsonFormatter()

_log = logging.getLogger()
_log.setLevel(logging.INFO)
_stream_handler = logging.StreamHandler()
_stream_handler.setFormatter(formatter)
_log.addHandler(_stream_handler)

info = _log.info
warning = _log.warning
error = _log.error
exception = _log.exception
