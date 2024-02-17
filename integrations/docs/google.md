# HOWTO: OAuth-Based Google APIs

This is an autokitteh integration for various Google APIs: Gmail, Calendar,
Sheets, etc.

## Prerequisites

1. FQDN of a public HTTPS tunnel to your autokitteh server

   - This is required for OAuth v2 flows and event webhooks

## GCP Project

1. Create a new GCP project:

   1. Follow the instructions at:
      <https://developers.google.com/workspace/guides/create-project>
   2. Quick link: <https://console.cloud.google.com/projectcreate>

2. Enable Google Workspace APIs

   1. Follow the instructions at:
      <https://developers.google.com/workspace/guides/enable-apis>
   2. Quick links:
      - Calendar:
        <https://console.cloud.google.com/apis/enableflow?apiid=calendar-json.googleapis.com>
      - Chat:
        <https://console.cloud.google.com/apis/enableflow?apiid=chat.googleapis.com>
      - Docs:
        <https://console.cloud.google.com/apis/enableflow?apiid=docs.googleapis.com>
      - Drive:
        <https://console.cloud.google.com/apis/enableflow?apiid=drive.googleapis.com>
      - Forms:
        <https://console.cloud.google.com/apis/enableflow?apiid=forms.googleapis.com>
      - Gmail:
        <https://console.cloud.google.com/apis/enableflow?apiid=gmail.googleapis.com>
      - Sheets:
        <https://console.cloud.google.com/apis/enableflow?apiid=sheets.googleapis.com>

## OAuth

1. Configure your OAuth consent page

   1. Follow the instructions at:
      <https://developers.google.com/workspace/guides/configure-oauth-consent>
   2. Quick link: <https://console.cloud.google.com/apis/credentials/consent>
      - Authorized domains:
        - See prerequisite 1
        - `localhost` - for testing, if allowed/desirable
   3. Add these scopes:
      - Non-sensitive:
        - `.../auth/userinfo.email`
        - `.../auth/userinfo.profile`
        - `openid`
      - Sensitive:
        - `.../auth/forms.body`
        - `.../auth/forms.responses.readonly`
        - `.../auth/spreadsheets`
      - Restricted:
        - `.../auth/drive`
        - `.../auth/gmail.modify`
        - `.../auth/gmail.settings.basic`

2. Create OAuth client ID credentials

   1. Follow the instructions at:
      <https://developers.google.com/workspace/guides/create-credentials#oauth-client-id>
   2. Quick link: <https://console.cloud.google.com/apis/credentials>
      - Click: `+ Create Credentials`
      - Select: `OAuth client ID`
      - Application type: `Web application`
      - Authorized redirect URI:
        - `https://<PREREQUISITE-1>/oauth/redirect/google`
        - `http://localhost:9980/oauth/redirect/google` - for testing, if allowed/desirable

## autokitteh Environment Variables

1. Create a `.env` file for autokitteh, if it doesn't already exist:

```shell
$ mkdir -m 700 ~/.autokitteh
$ touch ~/.autokitteh/.env
$ chmod 600 ~/.autokitteh/.env
```

2. Add these environment variables to the file `~/.autokitteh/.env`,
   based on the OAuth client above:

   - `GOOGLE_CLIENT_ID`
   - `GOOGLE_CLIENT_SECRET`
   - `WEBHOOK_ADDRESS`
     - See Prerequisite 1 above
     - Just the FQDN, no `https://` prefix, no path suffix
