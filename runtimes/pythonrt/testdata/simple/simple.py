# This is used by Test_pySvc_Run above, if you make any changes, make sure to run the
# test

from os import getenv
import json
import logging


def greet(event):
    logging.info("simple: HOME: %s", getenv('HOME'))  # From environment
    logging.info("simple: USER: %s", getenv('USER'))  # From "var" in manifest
    logging.info('simple: event: %r', event)

    body = event['data']['body']
    logging.info('BODY: %r', body)
    request = json.loads(body)
    logging.info('REQUEST: %r', request)
