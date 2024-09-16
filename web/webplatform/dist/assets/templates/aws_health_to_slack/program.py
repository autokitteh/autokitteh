"""Announce AWS health events in Slack, based on resource ownership in Google Sheet.

Documentation:
https://docs.aws.amazon.com/health/
https://aws.amazon.com/blogs/mt/tag/aws-health-api/
"""

from datetime import UTC, datetime, timedelta
import json
import os
import re

import autokitteh
from autokitteh.aws import boto3_client
from autokitteh.google import google_id, google_sheets_client
from autokitteh.slack import slack_client


url = os.getenv("GOOGLE_SHEET_URL")


def on_schedule(_):
    """Workflow's entry-point."""
    slack_channels = _read_google_sheet()
    events = _aws_health_events()
    events_by_arn = {event["arn"]: event for event in events}

    for entity in _affected_aws_entities(events):
        project = entity.get("tags", {}).get("project")
        if not project:
            print(f"Error: AWS entity without project tag: {entity}")
            continue

        channel = slack_channels.get(project)
        affecting_events = [events_by_arn[arn] for arn in entity["eventArns"]]
        _post_slack_message(project, channel, entity, affecting_events)


@autokitteh.activity
def _read_google_sheet() -> dict[str, str]:
    """Read mapping of project tags to Slack channels from Google Sheet."""
    sheets = google_sheets_client("google_sheets_connection").spreadsheets().values()
    rows = sheets.get(spreadsheetId=google_id(url), range="A:B").execute()
    return {row[0].strip(): row[1].strip() for row in rows.get("values", [])}


@autokitteh.activity
def _aws_health_events() -> list[dict]:
    """List all recent AWS Health events.

    This function currently fetches events for a single AWS account:
    https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/health/client/describe_events.html

    With a bit more code, you can also fetch events for multiple ones:
    https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/health/client/describe_events_for_organization.html
    """
    try:
        mins = int(re.match(r"(\d+)m", os.getenv("TRIGGER_INTERVAL")).group(1))
        prev_check = datetime.now(UTC) - timedelta(minutes=mins)
        filter = {"lastUpdatedTime": [{"from": prev_check}]}

        aws = boto3_client("aws_connection", "health")
        resp = aws.describe_events(filter=filter)
        events = resp.get("events", [])

        nextToken = resp.get("nextToken")
        while nextToken:
            resp = aws.describe_events(filter=filter, nextToken=nextToken)
            events += resp.get("events", [])
            nextToken = resp.get("nextToken")

        return events

    # TODO: More specific exception handling.
    except Exception as e:
        print(f"Error: {e}")
        return []


@autokitteh.activity
def _affected_aws_entities(events: list[dict]) -> list[dict]:
    """List all AWS entities affected by the given AWS Health events.

    API Reference:
    https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/health/client/describe_affected_entities.html
    """
    try:
        aws = boto3_client("aws_connection", "health")
        arns = [event["arn"] for event in events]
        # Possible alternative: describe_affected_entities_for_organization.
        resp = aws.describe_affected_entities(eventArn=arns)
        entities = resp.get("entities", [])

        nextToken = resp.get("nextToken")
        while nextToken:
            resp = aws.describe_affected_entities(eventArn=arns, nextToken=nextToken)
            entities += resp.get("entities", [])
            nextToken = resp.get("nextToken")

        return entities
    except Exception as e:
        print(f"Error: {e}")
        return []


def _post_slack_message(channel, project, entity: dict, affecting_events: list[dict]):
    if not channel:
        print(f"Error: project {project!r} not found in {url}")

    text = f"This AWS resource:\n```\n{json.dumps(entity, indent=4)}\n```"
    text += "\nis affected by these AWS Health events:"
    for i, event in enumerate(affecting_events, 1):
        text += f"\n{i}.\n```\n{json.dumps(event, indent=4)}\n```"

    print(f"Posting in Slack channel: {channel!r}")
    slack_client("slack_connection").chat_postMessage(channel=channel, text=text)
