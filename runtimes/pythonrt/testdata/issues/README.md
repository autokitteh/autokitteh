# issues Workflow

> [!WARNING]
> This code is work in progress.

When GitHub issue is opened, notify on slack.

## Setup

### GitHub

Create [a GitHub application][gh] and update `.env` in autokitteh config directory (use `ak config where` to find its location).

### Slack

Create a Slack application and update `SLACK_TOKEN` in `autokitteh.yaml` (the token should starts with `xoxb-`)
Update `SLACK_CHANNEL_ID` in autokitteh.yaml

## Running

Run temporal:

```
$ temporal server start-dev
```

Run autokitteh:

```
$ ak up -m dev
```

Create a ngrok tunnel:

```
$ ngrok http --domain autokitteh-miki.ngrok.dev 9980
```

- Open http://localhost:9980 and create a GitHub integration
- Update autokitteh.yaml with GitHub integration token
- Deploy the manifset (`ak deploy -m autokitteh.yaml -d .`)
- [Create a new issue][issue]
- A new message should appear in your slack channel


[issue]: https://github.com/testkitteh/miki/issues/
[gh]: https://docs.autokitteh.com/config/integrations/github
