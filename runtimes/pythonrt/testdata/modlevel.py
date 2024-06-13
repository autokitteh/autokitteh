from os import getenv
from random import randint

home = getenv('HOME')

ncalls = 0


def on_event(event):
    global ncalls

    ncalls += 1

    print('dice:', randint(1, 6))
