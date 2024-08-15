from os import getenv

HOME, USER = getenv("HOME"), getenv("USER")


def on_event(event):
    print(f"simple: HOME: {HOME}")  # From environment
    print(f"simple: USER: {USER}")  # From 'var' in manifest
    print(f"simple: event: {event!r}")
