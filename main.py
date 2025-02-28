#!/usr/bin/env ak worker --task-queue-name q

from requests import get
from time import sleep

from autokitteh import start, next_signal


def w0(event):
    print(event)
    return event


def w1(event):
    print(event)

    n, t = event.data.get("n", 1), event.data.get("t", 1)

    for i in range(n):
        sleep(t)

        print(get("https://example.com"))

    return {"t": n * t}


def w2(event):
    sid = start("main.py:w3", event.data)
    keys = next_signal(sid)
    return keys


def w3(event):
    return len(event.data)
