# Server Setup

## Introduction

This is the **first** autokitteh tutorial. It will guide you through the
procedure of setting up and starting a local server, in order to experiment
with autokitteh.

[The second tutorial](./cli_walkthrough.md) will demonstrate the deployment of
a project on this server, using the `ak` Command Line Interface.

[The third tutorial](./project_deep_dive.md) will be a deep dive into the
nitty-gritty details of projects, using low-level commands in the `ak` CLI
instead of the shortcuts featured in the second tutorial.

## Prerequisites

### Temporal

Install Temporal CLI, according to the platform-specific instructions at:
<https://learn.temporal.io/getting_started/go/dev_environment/>.

Open a new Terminal window and run the following command:

```console
temporal server start-dev --db-filename temporal.db --log-format pretty
```

This command starts a local Temporal Cluster, with a web UI and persistent
workflow state instead of an in-memory database.

Verification: the Temporal Web UI will be available at
<http://localhost:8233>.

See also: <https://docs.temporal.io/self-hosted-guide>.

### TODO: Optional Database (e.g. SQLite)

### TODO: Optional Secrets Manager (e.g. Vault)

## TODO: Install the AK Command Line Tool

<https://autokitteh.sh/>

Also mention shell autocomplete

## Start a Local Server

Run this command:

```console
ak up [--mode=dev]
```

> [!NOTE]
> You may append the flag `--mode=dev` to this command, which makes logs more
> human-readable, and relaxes some timeouts to make debugging easier.

## Conclusion

That's it!

You're ready for [the second tutorial](./cli_walkthrough.md), which will
demonstrate the deployment of a project on this server, using the `ak`
Command Line Interface.
