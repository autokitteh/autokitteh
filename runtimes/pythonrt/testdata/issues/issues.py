from os import getenv

from slack_sdk import WebClient

def format_message(issue):
    title, number, login, url = \
        issue['title'], issue['number'], issue['user']['login'], issue['html_url']

    return f'Issue #{number}: {title} opened by {login}, see {url}'


def on_issue(event):
    event = event['data']
    if event['action'] != 'opened':
        print(f'skipping event of type {event["action"]!r}')
        return

    issue = event['issue']
    print(f'issue: {issue}')
    text = format_message(issue)

    channel_id = getenv('SLACK_CHANNEL_ID')
    slack_token = getenv('SLACK_TOKEN')

    if not channel_id and slack_token:
        print('missing environment: SLACK_CHANNEL_ID, SLACK_TOKEN')

    client = WebClient(token=slack_token)
    client.chat_postMessage(channel=channel_id, text=text)
