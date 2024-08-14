"Bootstrap - contains common definition that are used in user code."

# All exported symbols must be declared using "# EXPORT:" comments.

# The followings are always supplied by the runtime: {
# EXPORT: run_activity
# EXPORT: catch
# EXPORT: fail
# EXPORT: globals
# EXPORT: module
# EXPORT: struct
# }
#
# TODO: Move the below to be supplied by the project build.
# The followings are always supplied by the session: {
# EXPORT: ak
# EXPORT: os
# EXPORT: rand
# EXPORT: store
# EXPORT: time
# }

# EXPORT: nop
def nop():
    pass

# EXPORT: poll
def poll(fn, pollerfn):
    orig = ak.poll(pollerfn)
    r = fn()
    ak.poll(orig)
    return r

# EXPORT: fake
def fake(*args, **kwargs):
    return ak.fake(*args, **kwargs)

# EXPORT: sleep
def sleep(*args, **kwargs):
    return ak.sleep(*args, **kwargs)

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
