# This is used by Test_pySvc_Run above, if you make any changes, make sure to run the
# test

from os import getenv
import json


def greet(event):
    print('INFO: simple: HOME:', getenv('HOME'))
    print('INFO: simple: USER:', getenv('USER'))
    print(f'INFO: simple: event: {event!r}')

    body = event['data']['body']
    print(f'BODY: {body!r}')
    request = json.loads(body)
    print(f'REQUEST: {request!r}')


if __name__ == '__main__':
    print(greet('garfield'))

