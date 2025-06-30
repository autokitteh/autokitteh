"""Making sure we run coroutines in the same event loop."""

import asyncio
from threading import Thread

_loop = None


def run_async(coro, timeout=None):
    """Run a coroutine in the AutoKitteh even loop."""
    global _loop

    if _loop is None:
        _loop = asyncio.new_event_loop()
        Thread(target=_loop.run_forever, daemon=True).start()

    fut = asyncio.run_coroutine_threadsafe(coro, _loop)
    return fut.result(timeout=timeout)
