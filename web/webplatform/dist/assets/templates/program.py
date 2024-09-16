"""
This program receives Jira events and creates Google Calendar events.

Scenario:
    Initiating a procedure that requires collaboration and coordination,
    e.g. scheduling a consult with another team, or planning a joint review.

Workflow:
    The user creates a new Jira ticket for the discussion. AutoKitteh
    automatically generates a Google Calendar event with a deadline for
    the completion, to ensure that the review happens as planned.
"""

import autokitteh
from autokitteh import atlassian
from autokitteh.google import google_calendar_client


def on_jira_issue_created(event):
    """Workflow's entry-point."""
    _create_calendar_event(event.data.issue.fields, event.data.issue.key)


@autokitteh.activity
def _create_calendar_event(issue, key):
    url = atlassian.get_base_url("jira_connection")
    link = f"Link to Jira issue: {url}/browse/{key}\n\n"

    event = {
        "summary": issue.summary,
        "description": link + issue.description,
        "start": {"date": issue.duedate},
        "end": {"date": issue.duedate},
        "reminders": {"useDefault": True},
        "attendees": [
            {"email": "auto@example.com"},
            {"email": "kitteh@example.com"},
        ],
    }

    gcal = google_calendar_client("google_calendar_connection").events()
    event = gcal.insert(calendarId="primary", body=event).execute()

    print("Google Cloud event created: " + event.get("htmlLink"))
