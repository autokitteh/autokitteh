# issues Workflow

**WARNING: WIP**

When GitHub issue is opened, notify on slack.

## Running

Run temporal:

```
$ temporal server start-dev
```

Run autokitteh:

```
$ ak up -m dev
```

Create an ngrok tunnel:

```
$ ngrok http --domain autokitteh-miki.ngrok.dev 9980
```

- Deploy the manifset (see ../deploy.go)
- [Create a new issue][issue]
- A new message should appear in `#slack-test` channel on autokitteh Slack



[issue]: https://github.com/testkitteh/miki/issues/
