"""asyncio interface for AutoKitteh SDK"""


# Hack to add async version of everything in autokitteh API
def _init():
    import autokitteh
    from functools import wraps

    def make_wrapper(fn):
        @wraps(fn)
        async def wrapper(*args, **kw):
            return fn(*args, **kw)

        return wrapper

    env = globals()
    for name in autokitteh.__all__:
        fn = getattr(autokitteh, name)
        # FIXME: Store is a dict that under the hood calls gRPC methods, don't expose it ATM
        if name == "store":
            continue

        if not callable(fn):
            continue

        env[name] = make_wrapper(fn)


_init()
del _init
