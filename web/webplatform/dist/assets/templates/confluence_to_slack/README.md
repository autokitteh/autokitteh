# Confluence To Slack Workflow 

This real-life workflow demonstrates the integration between two popular services.

## Benefits

- **Small overhead**: Run the `ak` server, deploy the project, and write code.
- **Filtering**: Add filters in the configuration to limit the number of times your code is triggered, or filter data in the code itself. This workflow demonstrates both.

## How It Works

- **Trigger**: A new Confluence page is created in a designated space.
- **Result**: A Slack message is sent to a selected channel containing data from the newly created Confluence page.

## Known Limitations

- Confluence returns HTML, and this program does not format it in any way. The purpose of this workflow is to demonstrate how data can move between different services. Desired formatting can be easily added to suit individual needs.

## Additional Comment

- Environment variables are set in [`autokitteh.yaml`](./autokitteh.yaml) (e.g., Slack channel, Confluence page, etc.).
