# HOWTO: GitHub App for autokitteh

## Prerequisites

1. FQDN of a public HTTPS tunnel to your autokitteh server \
   (autokitteh-internal convention: `autokitteh-<USERNAME>.ngrok.dev`)

   - This is required for OAuth v2 flows and event webhooks

2. You need to be a GitHub organization **owner** in order to install and
   update the settings of the app
   (autokitteh-internal convention: <https://github.com/testkitteh>)

## App Creation

App creation is different from - but related to - app installation:

- If you want the app to be private, you have to create it in the same
  org/user scope where you intend to install it.
- If you want to install it for multiple orgs/users, you have to create
  either one public app or multiple private duplicate apps.

Choose one of these options:

1. The app to be owned by a GitHub organization:
   <https://github.com/organizations/ORG-NAME/settings/apps>

2. The app to be owned by you, i.e. a GitHub user:
   <https://github.com/settings/apps>

Either way, click the "New GitHub App" button in the top-right corner
of the app settings page.

## App Settings

- GitHub App Name \
  (autokitteh-internal convention: `autokitteh-<USERNAME>`)
- Homepage URL \
  (autokitteh-internal convention: `https://autokitteh.com/`)
- Callback URL: `https://<PREREQUISITE-1>/oauth/redirect/github`
- **(TEMPORARY)** Expire user authorization tokens: **No**
- Request user authorization (OAuth) during installation: **Yes**
- Post installation - Redirect on update: **Yes**
- Webhook URL: `https://<PREREQUISITE-1>/github/webhook` \
  (not the same as the callback URL above!)
- Webhook Secret: random and secret string \
  (autokitteh-internal convention for dev & test: `<USERNAME>`)
- Repository permissions:
  - [Actions](https://docs.github.com/en/rest/overview/permissions-required-for-github-apps?apiVersion=2022-11-28#repository-permissions-for-actions) (read and write)
  - [Administration](https://docs.github.com/en/rest/overview/permissions-required-for-github-apps?apiVersion=2022-11-28#repository-permissions-for-administration) (read-only)
  - [Commit statuses](https://docs.github.com/en/rest/overview/permissions-required-for-github-apps?apiVersion=2022-11-28#repository-permissions-for-commit-statuses) (read and write)
  - [Contents](https://docs.github.com/en/rest/overview/permissions-required-for-github-apps?apiVersion=2022-11-28#repository-permissions-for-contents) (read-only)
  - [Issues](https://docs.github.com/en/rest/overview/permissions-required-for-github-apps?apiVersion=2022-11-28#repository-permissions-for-issues) (read and write)
  - [Metadata](https://docs.github.com/en/rest/overview/permissions-required-for-github-apps?apiVersion=2022-11-28#repository-permissions-for-metadata) (read-only)
  - [Pull requests](https://docs.github.com/en/rest/overview/permissions-required-for-github-apps?apiVersion=2022-11-28#repository-permissions-for-pull-requests) (read and write)
- Subscribe to events:
  - [Commit comment](https://docs.github.com/en/webhooks/webhook-events-and-payloads#commit_comment)
  - [Issue comment](https://docs.github.com/en/webhooks/webhook-events-and-payloads#issue_comment)
  - [Issues](https://docs.github.com/en/webhooks/webhook-events-and-payloads#issues)
  - [Meta](https://docs.github.com/en/webhooks/webhook-events-and-payloads#meta)
  - [Pull request](https://docs.github.com/en/webhooks/webhook-events-and-payloads#pull_request)
  - [Pull request review](https://docs.github.com/en/webhooks/webhook-events-and-payloads#pull_request_review)
  - [Pull request review comment](https://docs.github.com/en/webhooks/webhook-events-and-payloads#pull_request_review_comment)
  - [Pull request review thread](https://docs.github.com/en/webhooks/webhook-events-and-payloads#pull_request_review_thread)
  - [Repository](https://docs.github.com/en/webhooks/webhook-events-and-payloads#repository)
  - [Status](https://docs.github.com/en/webhooks/webhook-events-and-payloads#status)
  - [Workflow job](https://docs.github.com/en/webhooks/webhook-events-and-payloads#workflow_job)
  - [Workflow run](https://docs.github.com/en/webhooks/webhook-events-and-payloads#workflow_run)

Where can this GitHub App be installed?

- `Only on this account` (the org/user that creates the app)
- `Any account` (any org/user in GitHub)

And finally, click the "Create GitHub App" button.

## App Secrets

1. Click the "generate a new client secret" button at the **top** of the app
   settings page, and copy it

2. Click the "Generate a private key" button at the **bottom** of the app
   settings page

   - This will auto-download a file named `NAME.DATE.private-key.pem`
   - Convert this file into a single string with this command-line:
     ```shell
     cat autokitteh.YY-MM-DD.private-key.pem | awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}'
     ```
   - Delete this file

## autokitteh Environment Variables

1. Create a `.env` file for autokitteh, if it doesn't already exist:

```shell
$ mkdir -m 700 ~/.autokitteh
$ touch ~/.autokitteh/.env
$ chmod 600 ~/.autokitteh/.env
```

2. Add these environment variables to the file `~/.autokitteh/.env`,
   based on the app settings and secrets:

   - `GITHUB_APP_NAME`
   - `GITHUB_CLIENT_ID`
   - `GITHUB_CLIENT_SECRET`
     - Readbale only when generated
   - `GITHUB_PRIVATE_KEY`
     - Downloadable only when generated
   - `GITHUB_WEBHOOK_SECRET`
     - Readbale only when re/set
   - `WEBHOOK_ADDRESS`
     - See Prerequisite 1 above
     - Just the FQDN, no `https://` prefix, no path suffix
     - Used for the app's Callback and Webhook URLs
