import autokitteh


@autokitteh.activity
async def on_event(event):
    print("on_event: start")
    out = await work()
    print(f"on_event: end ({out=})")


async def work():
    print("work")
    return 8


if __name__ == "__main__":
    import asyncio

    asyncio.run(on_event(None))
