# This is used by Test_pySvc_Run above, if you make any changes, make sure to run the
# test

import json
from os import getenv

import autokitteh
from autokitteh.connections import refresh_oauth

HOME, USER = getenv("HOME"), getenv("USER")


def greet(event):
    refresh_oauth("lassie")

    print(f"simple: HOME: {HOME}")  # From environment
    print(f"simple: USER: {USER}")  # From 'var' in manifest
    print(f"simple: event: {event!r}")

    body = event.data.body.bytes
    print(f"BODY: {body!r}")
    request = json.loads(body)
    print(f"REQUEST: {request!r}")


@autokitteh.activity
def printer(event):
    print(event)
