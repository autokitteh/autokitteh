# Pull Request Review Reminder (Purrr)

Purrr integrates GitHub and Slack efficiently, to streamline code reviews and
cut down the turnaround time to merge pull requests.

- Real-time, relevant, informative, 2-way updates
- Easier collaboration and faster execution
- Free and open-source

No more:

- Delays due to missed requests, comments, and state changes
- Notification fatigue due to updates that don't concern you
- Qestions like "Who's turn is it" or "What should I do now"

All that and more is implemented with a few hundred lines of Python.
AutoKitteh takes care of all the system infrastructure and reliability needs.

## Slack Usage

Event-based, 2-way synchronization:

- Slack channels are created and archived automatically for each PR
- Stakeholders are added and removed automatically in Slack and GitHub
- Reviews, comments, conversation threads, and emoji reactions are updated in
  both directions

User matching between GitHub and Slack is based on email addresses and
case-insensitive full names.

Available Slack slash commands:

- `/ak purrr help`
- `/ak purrr opt-in`
- `/ak purrr opt-out`
- `/ak purrr list`
- `/ak purrr status [PR]`
- `/ak purrr approve [PR]`

## Data Storage

Purrr uses Redis or Valkey for:

1. Mapping between GitHub PRs and Slack channels
2. Mapping between GitHub comments/reviews and Slack message threads
3. Caching user IDs (optimization to reduce API calls)
4. User opt-out database

Use-cases 1 and 2 use a TTL of 30 days (configurable in the `autokitteh.yaml`
manifest file). Use-case 3 uses a TTL of one day since the last cache hit.
Use-case 4 is permanent (until the user opts back in).
