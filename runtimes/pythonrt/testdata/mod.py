# Test file for loader

import json
from datetime import datetime


# Internal function, shouldn't be patched
def log(msg):
    now = datetime.now()
    print(f'[{now}] msg')


def parse(data):
    log(f'parsing {data!r}')
    # External function, should be patched
    return json.loads(data)
