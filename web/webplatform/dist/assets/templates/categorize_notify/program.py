"""
This program demonstrates a real-life workflow that integrates Gmail, ChatGPT, and Slack.

Workflow:
1. Trigger: Detect a new email in Gmail.
2. Categorize: Use ChatGPT to read and categorize the email (e.g., technical work, marketing, sales).
3. Notify: Send a message to the corresponding Slack channel based on the category.
4. Label: Add a label to the email in Gmail.
"""

import base64
import time

import autokitteh
from autokitteh import google, openai, slack


SLACK_CHANNELS = ["demos", "engineering", "ui"]


def on_http_get(event):
    total_messages = None
    while True:
        total_messages = _poll_inbox(total_messages)
        time.sleep(10)


@autokitteh.activity
def _poll_inbox(prev_total: int):
    gmail = google.gmail_client("my_gmail").users()
    curr_total = gmail.getProfile(userId="me").execute()["messagesTotal"]
    # Note: This is not meant to be a robust solution for handling email operations.
    if prev_total and curr_total > prev_total:
        new_email_count = curr_total - prev_total
        message_ids = _get_latest_message_ids(gmail, new_email_count)
        for message_id in message_ids:
            _process_email(gmail, message_id)

    return curr_total


def _process_email(gmail, message_id: str):
    message = gmail.messages().get(userId="me", id=message_id).execute()
    email_content = _parse_email(message)
    if email_content:
        channel = _categorize_email(email_content)
        if channel:
            client = slack.slack_client("my_slack")
            client.chat_postMessage(channel=channel, text=email_content)

        # Add label to email
        label_id = _get_label_id(gmail, channel) or _create_label(gmail, channel)
        body = {"addLabelIds": [label_id]}
        gmail.messages().modify(userId="me", id=message_id, body=body).execute()


def _get_latest_message_ids(gmail, new_email_count: int):
    results = gmail.messages().list(userId="me", maxResults=new_email_count).execute()
    return [msg["id"] for msg in results.get("messages", [])]


def _parse_email(message: dict):
    payload = message["payload"]
    for part in payload.get("parts", []):
        if part["mimeType"] == "text/plain":
            return base64.urlsafe_b64decode(part["body"]["data"]).decode("utf-8")


def _create_label(gmail, label_name: str) -> str:
    """Create a new label in the user's gmail account.

    https://developers.google.com/gmail/api/reference/rest/v1/users.labels#Label
    """
    label = {
        "labelListVisibility": "labelShow",
        "messageListVisibility": "show",
        "name": label_name,
    }
    created_label = gmail.labels().create(userId="me", body=label).execute()
    print(f"Label created: {created_label['name']}")
    return created_label["id"]


def _get_label_id(gmail, label_name: str) -> str:
    labels_response = gmail.labels().list(userId="me").execute()
    labels = labels_response.get("labels", [])
    for label in labels:
        if label["name"] == label_name:
            return label["id"]
    return None


@autokitteh.activity
def _categorize_email(email_content: str) -> str:
    """Prompt ChatGPT to categorize an email based on its content.

    Returns:
        The name of the Slack channel to send a message to as a string.
        If the channel is not in the provided list, returns None.
    """
    client = openai.openai_client("my_chatgpt")
    response = client.chat.completions.create(
        model="gpt-3.5-turbo",
        messages=[
            {"role": "system", "content": "You are a helpful assistant."},
            {
                "role": "user",
                "content": f"""Categorize the following email based on its
                topic and suggest a channel to post it in from the 
                provided list. The output should be one of the provided 
                channels and nothing else.
                Email Content: {email_content} Channels: {SLACK_CHANNELS}
                Output example: {SLACK_CHANNELS[0]}""",
            },
        ],
    )
    channel = response.choices[0].message.content
    return channel if channel in SLACK_CHANNELS else None
