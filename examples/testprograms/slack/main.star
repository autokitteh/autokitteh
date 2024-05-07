load("@slack", "slack")

def on_slack_app_mention(data):
    slack.chat_post_message(data.channel, "meow")
