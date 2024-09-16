# JIRA Assignee From Google Calendar Workflow 

This real-life example workflow demonstrates the integration of JIRA and Google Calendar in an on-call scenario.

## Benefits

- **Focus on what matters**: Write code that focuses on the desired outcome, not the underlying infrastructure.
- **Flexibility**: Implement your own authorization flow or use the one that works out of the box.
- **Extensibility**: Easily add additional steps or integrations.

## How It Works

- **Trigger**: A new Jira issue in the designated Jira project (specified in [`autokitteh.yaml`](./autokitteh.yaml))
- **Result**: The current person on-call is retrieved via the Google Calendar API and added as the assignee in the Jira issue that triggered the workflow.
