"""See https://autokitteh.readthedocs.io/en/latest/"""

import os

from autokitteh.github import github_client
from autokitteh.google import gmail_client
from autokitteh.google import google_calendar_client
from autokitteh.google import google_forms_client
from autokitteh.slack import slack_client

import print


def all(event):
    github_get_repo("GitHub repo", event)
    gmail_get_profile(event)
    google_calendar_list(event)
    google_forms_get(event)
    slack_auth_test(event)


def github_get_repo(_):
    github = github_client("github_conn")
    repo = github.get_repo("autokitteh/autokitteh")
    print("GitHub:", repo)


def gmail_get_profile(_):
    gmail = gmail_client("gmail_conn").users()
    profile = gmail.getProfile(userId="me").execute()
    print.pretty_json("Gmail profile", profile)


def google_calendar_list(_):
    calendar_id = os.getenv("calendar_conn__CalendarID")
    if not calendar_id:
        print("No Google Calendar is being watched")
    else:
        print("Watched Google Calendar ID:", calendar_id)

    gcal = google_calendar_client("calendar_conn")
    req = gcal.calendarList().list()
    while req:
        resp = req.execute()
        for item in resp["items"]:
            print.pretty_json("Google Calendar", item)
        req = gcal.calendarList().list_next(req, resp)


def google_forms_get(_):
    form_id = os.getenv("forms_conn__FormID")
    if not form_id:
        print("No Google Form is being watched, can't make API call!")
    else:
        forms = google_forms_client("forms_conn").forms()
        form = forms.get(formId=form_id).execute()
        print.pretty_json("Google Form", form)


def slack_auth_test(_):
    slack = slack_client("slack_conn")
    print.pretty_json("Slack auth", slack.auth_test().data)
