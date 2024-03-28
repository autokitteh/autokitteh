# Test file for loader

import json


# Internal function, shouldn't be patched
def log(msg):
    print(msg)


def parse(data):
    log(f'parsing {data!r}')
    # External function, should be patched
    return json.loads(data)
