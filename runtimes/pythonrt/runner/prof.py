import os

import psutil

import log

_process = psutil.Process(os.getpid())
_peak_rss = 0


def sample() -> psutil.pmem:
    global _peak_rss
    curr = _process.memory_info()
    if curr.rss > _peak_rss:
        _peak_rss = curr.rss
    log.debug("memory sample: rss=%d, peak=%d", curr.rss, _peak_rss)
    return curr


def report() -> None:
    log.info("peak memory usage: rss=%d bytes", _peak_rss)
