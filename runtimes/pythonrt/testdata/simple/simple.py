# This is used by Test_pySvc_Run above, if you make any changes, make sure to run the
# test

from os import getenv
import json


def greet(event):
    home, user = getenv('HOME'), getenv('USER')
    print(f'simple: HOME: {home}')  # From environment
    print(f'simple: USER: {user}')  # From 'var' in manifest
    print(f'simple: event: {event!r}')

    body = event['data']['body']
    print(f'BODY: {body!r}')
    request = json.loads(body)
    print(f'REQUEST: {request!r}')
