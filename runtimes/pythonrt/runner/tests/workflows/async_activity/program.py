import autokitteh


async def on_event(event):
    await work()


@autokitteh.activity
async def work():
    pass
