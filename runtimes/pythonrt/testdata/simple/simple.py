# This is used by Test_pySvc_Run above, if you make any changes, make sure to run the
# test

import json
from os import getenv
from time import sleep
from datetime import datetime

import autokitteh

HOME, USER = getenv("HOME"), getenv("USER")


def greet(event):
    print(f"simple: HOME: {HOME}")  # From environment
    print(f"simple: USER: {USER}")  # From 'var' in manifest
    print(f"simple: event: {event!r}")

    body = event["data"]["body"]
    print(f"BODY: {body!r}")
    request = json.loads(body)
    print(f"REQUEST: {request!r}")

    print("NOW:", datetime.now())
    sleep(1)
    print("NOW:", datetime.now())


@autokitteh.activity
def printer(event):
    print(event)
