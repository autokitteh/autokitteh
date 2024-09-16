# Email Categorization and Notification Workflow
This project automates the process of categorizing incoming emails and notifying relevant Slack channels by integrating Gmail, ChatGPT, and Slack. It is not meant to be a 100% complete project, but rather a solid starting point.

## Benefits
- **Ease of Use:** Demonstrates how easy it is to connect multiple integrations into a cohesive workflow.
- **Low Complexity:** The workflow is implemented with a minimal amount of code.
- **Free and Open Source:** Available for use or modification to fit specific use cases. 

## How It Works
- **Detect New Email**: The program monitors the Gmail inbox for new emails using the Gmail API.
- **Categorize Email**: ChatGPT analyzes the email content and categorizes it into predefined categories.
- **Send Slack Notification**: The program sends the categorized email content to the corresponding Slack channel using the Slack API.
- **Label Email**: Adds a label to the processed email in Gmail for tracking.

For more details, refer to [this blog post](https://autokitteh.com/technical-blog/from-inbox-to-slack-automating-email-categorization-and-notifications-with-ai/).

## Known Limitations
- **ChatGPT**: the prompt for this workflow works for simple cases. It may have mixed results for emails that lack detail or have nothing to do with the channels provided.
- **E-mail polling**: the polling mechanism is basic and does not cover edge cases.
