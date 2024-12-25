# This is used by Test_pySvc_Run above, if you make any changes, make sure to run the
# test

import json
from os import getenv

import autokitteh
from printer import display

HOME, USER = getenv("HOME"), getenv("USER")


def greet(event):
    display(f"simple: HOME: {HOME}")  # From environment
    display(f"simple: USER: {USER}")  # From 'var' in manifest
    try:
        printer(event)
    except Exception as err:
        print("ZERR:", err, type(err))

    body = event.data.body.bytes
    display(f"BODY: {body!r}")
    request = json.loads(body)
    display(f"REQUEST: {request!r}")


@autokitteh.activity
def printer(event):
    print(f"simple: event: {event!r}")
    1 / 0
