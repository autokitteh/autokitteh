import logging

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(filename)s:%(lineno)d - %(message)s",
    datefmt="%Y-%m-%dT%H:%M:%S",
)

_log = logging.getLogger("runner")

info = _log.info
warning = _log.warning
error = _log.error
exception = _log.exception
