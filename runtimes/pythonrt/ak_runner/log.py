import logging

_log = logging.getLogger("AK")

info = _log.info
warning = _log.warning
error = _log.error
exception = _log.exception


class AKLogHandler(logging.Handler):
    def __init__(self, level, comm):
        super().__init__(level)
        self.comm = comm
        self.formatter = logging.Formatter()

    def emit(self, record):
        level = "ERROR" if record.levelname == "CRITICAL" else record.levelname
        message = record.getMessage()
        if record.exc_info:
            message += "\n" + self.formatter.formatException(record.exc_info)
        self.comm.send_log(level, message)


def init(level, comm):
    _log.setLevel(level)
    handler = AKLogHandler(level, comm)
    _log.addHandler(handler)
