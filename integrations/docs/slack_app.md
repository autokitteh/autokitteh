# HOWTO: Slack App for autokitteh

## Prerequisites

1. FQDN of a public HTTPS tunnel to your autokitteh server \
   (autokitteh-internal convention: `autokitteh-<USERNAME>.ngrok.dev`)

   - This is required for OAuth v2 flows and event webhooks

## App Creation

1. Go to <https://api.slack.com/apps>, and click the "Create New App" button

2. Select the option "From an app manifest"

3. Pick a workspace to develop the app in

4. Paste the YAML manifest below
   - Replace all `<TODO>` instances with actual content \
     (autokitteh-internal convention: `ak-<USERNAME>`)
   - Replace all `<PREREQUISITE-1>` instances with the FQDN

```yaml
display_information:
  name: <TODO>
  description: <TODO>
features:
  bot_user:
    display_name: <TODO>
    always_online: true
  slash_commands:
    - command: /<TODO>
      url: https://<PREREQUISITE-1>/slack/command
      description: Send command to autokitteh
      usage_hint: help
      should_escape: true
oauth_config:
  redirect_urls:
    - https://<PREREQUISITE-1>/oauth/redirect/slack
  scopes:
    bot:
      - app_mentions:read
      - bookmarks:read
      - bookmarks:write
      - channels:history
      - channels:manage
      - channels:read
      - chat:write
      - chat:write.public
      - commands
      - dnd:read
      - groups:history
      - groups:read
      - groups:write
      - im:history
      - im:read
      - im:write
      - mpim:history
      - mpim:read
      - mpim:write
      - reactions:read
      - reactions:write
      - users.profile:read
      - users:read
      - users:read.email
settings:
  event_subscriptions:
    request_url: https://<PREREQUISITE-1>/slack/event
    bot_events:
      - app_mention
      - app_uninstalled
      - channel_archive
      - channel_created
      - channel_deleted
      - channel_history_changed
      - channel_id_changed
      - channel_left
      - channel_rename
      - channel_unarchive
      - group_archive
      - group_deleted
      - group_history_changed
      - group_left
      - group_open
      - group_rename
      - group_unarchive
      - im_history_changed
      - member_joined_channel
      - message.channels
      - message.groups
      - message.im
      - message.mpim
      - reaction_added
      - reaction_removed
      - tokens_revoked
  interactivity:
    is_enabled: true
    request_url: https://<PREREQUISITE-1>/slack/interaction
  org_deploy_enabled: false
  socket_mode_enabled: false
  token_rotation_enabled: false
```

5. Click the "Create" button to create the app

6. Click the "Install to Workspace" button, and then the "Allow" button

7. Go to the "App Home" settings page

   - Show Tabs > Message Tab > Allow users to send Slash commands and messages
     from the messages tab

**Attention:** If you created the app when the autokitteh server wasn't up,
you may need to go to the Slack app's "Event Subscriptions" page and retry
validating the webhook URL while the autokitteh server is up and ready to
respond to URL verification requests from Slack.

## autokitteh Environment Variables

1. Create a `.env` file for autokitteh, if it doesn't already exist:

```shell
$ mkdir -m 700 ~/.autokitteh
$ touch ~/.autokitteh/.env
$ chmod 600 ~/.autokitteh/.env
```

2. Add these environment variables to the file `~/.autokitteh/.env`,
   based on the app settings and secrets:

   - `SLACK_APP_ID`
   - `SLACK_CLIENT_ID`
   - `SLACK_CLIENT_SECRET`
   - `SLACK_SIGNING_SECRET`
   - `WEBHOOK_ADDRESS`
     - See Prerequisite 1 above
     - Just the FQDN, no `https://` prefix, no path suffix
     - Used for the app's OAuth redirect URL, and webhook URLs
       ("Interactivity & Shortcuts" and "Event Subscriptions")
