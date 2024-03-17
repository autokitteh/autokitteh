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

def on_issue(event):
    if event['action'] != 'opened':
        logging.info('skipping event of type %r', event['action'])
        return

    issue = event['issue']
    title, number, login, url = \
        issue['title'], issue['number'], issue['user']['login'], issue['url']

    logging.info('issue %s: %s by %s', number, title, login)
    text = {
        'type': 'mrkdwn',
        'text': f'<{url}|Issue #{number}>: {title} opened by {login}',
    }

    client = WebClient(token=SLACK_TOKEN)
    client.chat_post_message(channel=CHANNEL_ID, text=text)


if __name__ == '__main__':
    event = {
        'action': 'opened',
        'issue': {
            'title': 'Fix url',
            'number': 1,
            'user': {'login': 'tebeka'},
            'url': 'https://api.github.com/repos/tebeka/toggl/issues/1',
        }
    }

    on_issue(event)
