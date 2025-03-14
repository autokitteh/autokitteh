"Bootstrap - contains common definition that are used in user code."

# All exported symbols must be declared using "# EXPORT:" comments.

# The followings are always supplied by the runtime: {
# EXPORT: ak
# EXPORT: catch
# EXPORT: fail
# EXPORT: globals
# EXPORT: module
# EXPORT: rand
# EXPORT: run_activity
# EXPORT: sleep
# EXPORT: struct
# }

# EXPORT: nop
def nop():
    pass

# EXPORT: start
def start(*args, **kwargs):
    return ak.start(*args, **kwargs)

# EXPORT: subscribe
def subscribe(*args, **kwargs):
    return ak.subscribe(*args, **kwargs)    

# EXPORT: next_event
def next_event(*args, **kwargs):
    return ak.next_event(*args, **kwargs)

# EXPORT: unsubscribe
def unsubscribe(*args, **kwargs):
    return ak.unsubscribe(*args, **kwargs)

# EXPORT: next_signal
def next_signal(*args, **kwargs):
    return ak.next_signal(*args, **kwargs)

# EXPORT: signal
def signal(*args, **kwargs):
    return ak.signal(*args, **kwargs)

# EXPORT: test
def test():
    for name, fn in globals().items():
        if not name.startswith("test_"):
            continue

        print("TEST: {}".format(name))
        fn()
        # TODO: report only errors in this step.

# EXPORT: activity
def activity(fn):
    return struct(
        run=lambda *args, **kwargs: run_activity(fn, *args, **kwargs),
    )
