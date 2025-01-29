from time import sleep
import time


def handler(event):
    print("before")
    sleep(1)
    time.sleep(2)
    print("after")
