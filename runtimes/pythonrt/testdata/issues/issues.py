import logging
import os

from slack_sdk import WebClient

CHANNEL_ID = os.getenv('SLACK_CHANNEL_ID')
SLACK_TOKEN = os.getenv('SLACK_TOKEN')


logging.basicConfig(
    format='%(asctime)s - %(levelname)s - %(filename)s:%(lineno)d - %(message)s',
    datefmt='%Y-%M-%DT%H:%M:%S',
    level=logging.INFO,
)


def format_message(issue):
    title, number, login, url = \
        issue['title'], issue['number'], issue['user']['login'], issue['html_url']

    return f'Issue #{number}: {title} opened by {login}, see {url}'


def on_issue(event):
    event = event['data']
    if event['action'] != 'opened':
        logging.info('skipping event of type %r', event['action'])
        return

    issue = event['issue']
    logging.info('issue %r:', issue)
    text = format_message(issue)

    client = WebClient(token=SLACK_TOKEN)
    client.chat_postMessage(channel=CHANNEL_ID, text=text)
