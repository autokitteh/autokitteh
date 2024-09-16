"""
This program assignes Atlassian Jira issues based on a shared Google Calendar.

The shared Google Calendar defines a 27/4 on-call rotation.
How to create it: https://support.google.com/calendar/answer/37095

This program assumes that the calendar entries have these fields:
- Summary: the on-call person's human-readable name
- Description: their Atlassian account ID
"""

from datetime import UTC, datetime, timedelta
import os

import autokitteh
from autokitteh.atlassian import atlassian_jira_client
from autokitteh.google import google_calendar_client


def on_jira_issue_created(event):
    """Workflow's entry-point."""
    name, account_id = _get_current_oncall()
    update = {"assignee": {"accountId": account_id}}

    jira = atlassian_jira_client("jira_connection")
    jira.update_issue_field(event.data.issue.key, update, notify_users=True)

    print(f"Assigned {event.data.issue.key} to {name}")


@autokitteh.activity
def _get_current_oncall():
    """Return the name and Atlassian account ID of the current on-call."""
    gcal = google_calendar_client("google_calendar_connection").events()
    now = datetime.now(UTC)
    in_a_minute = now + timedelta(minutes=1)

    result = gcal.list(
        calendarId=os.getenv("SHARED_CALENDAR_ID"),
        timeMin=now.isoformat(),  # Request all currently-effective events.
        timeMax=in_a_minute.isoformat(),
        orderBy="updated",  # Use the most-recently updated one.
    ).execute()["items"][-1]

    # Google Calendar may add whitespaces - strip them.
    return result["summary"].strip(), result["description"].strip()
