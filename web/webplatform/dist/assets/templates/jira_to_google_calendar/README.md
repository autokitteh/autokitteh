# Jira to Google Calendar Workflow

This program is a real-life example workflow integrating Jira and Google Calendar.

## Benefits

- **Simplicity**: In a few lines of code, you have a functional workflow that is authenticated and integrated, allowing two applications to communicate with each other.
- **Flexibility**: Use this as a starting point. Add integrations or change the trigger. It's open source and free, so there are no limits to what you can do.

## How It Works

- **Trigger**: The creation of a Jira issue in the designated project (specified in the [`autokitteh.yaml`](./autokitteh.yaml) file).
- **Result**: A new event is created in the user's Google Calendar containing information from the Jira issue (e.g., `duedate`).

## Known Limitations

- Attendees are hard-coded and arbitrary.
- Error handling is not implemented for demo purposes. For example, if the Jira issue is missing any of the required fields (e.g., `description`, `duedate`), the program will not fail gracefully.
