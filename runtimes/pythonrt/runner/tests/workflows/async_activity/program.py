import autokitteh


async def on_event(event):
    print("on_event: start")
    out = await work()
    print(f"on_event: end ({out=})")


@autokitteh.activity
async def work():
    print("work")
    out = await value()
    return out


async def value():
    return 8


if __name__ == "__main__":
    import asyncio

    asyncio.run(on_event(None))
