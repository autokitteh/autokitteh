import os
from random import choice
from time import sleep

from autokitteh import github, google, slack

SHEET_ID = '1W3T-OQd-X9aT021J-8Lqh_CvaWCfPTasy1tdHGzRgjY'
CHANNEL = os.getenv('CHANNEL')


def on_github_pull_request(data):
    if data.action not in ['opened', 'reopened']:
        return

    pr = data.pull_request
    href = pr.links.html.h_ref
    ts = slack.chat_post_message(CHANNEL, f'{href} [{pr.state}]').ts

    i = 0
    login, repo = data.repo.owner.login, data.repo.name
    while pr.state not in ['closed', 'merged']:
        sleep(1)

        pr = github.get_pull_request(login, repo, pr.number)

        href = pr.links.html.h_ref
        slack.chat_update(CHANNEL, ts, f'{href} meow [{pr.state}]')

        i += 1
        if i % 3 == 0:
            rows = google.sheets.read_range(SHEET_ID, 'A1:A5')
            the_chosen_one=choice(rows)
            email =f'{the_chosen_one}@autokitteh.com'
            user = slack.users_lookup_by_email(email).user
            slack.chat_post_message(CHANNEL, f'paging <@{user.id}>', thread_ts=ts)
