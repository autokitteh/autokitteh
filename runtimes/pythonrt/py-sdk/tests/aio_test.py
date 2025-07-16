import asyncio

from autokitteh import aio

loop1 = loop2 = None


async def garfield():
    global loop1

    loop1 = asyncio.get_running_loop()
    return "Garfield"


async def odie():
    global loop2

    loop2 = asyncio.get_running_loop()
    return "Odie"


def test_run_async():
    v1 = aio.run_async(garfield())
    assert v1 == "Garfield"

    v2 = aio.run_async(odie())
    assert v2 == "Odie"

    assert loop1 is loop2
