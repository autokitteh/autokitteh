from datetime import datetime
from time import sleep


def on_event(event):
    print(f"EVENT: {event!r}")

    now = datetime.now()
    print(f"NOW: {now}")
    sleep(10 * 60)
