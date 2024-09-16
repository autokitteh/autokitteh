"""A program that listens for GitHub pull requests and meows at random people.

This program listens for GitHub pull request events and posts a message to a
Slack channel when a pull request is opened or reopened. It then polls the
pull request until it is closed or merged, updating the message with the
current state of the pull request. Every 15 seconds, it also reads a random
name from a Google Sheet and pages that person in the Slack channel.
"""

load("@slack", "my_slack")
load("@github", "my_github")
load("@googlesheets", "my_googlesheets")
load("env", "CHANNEL", "SHEET_ID")


def on_github_pull_request(data):
    """Workflow's entry-point."""
    if data.action not in ["opened", "reopened"]:
        return

    pr = data.pull_request
    message = "%s [%s]" % (pr.links.html.h_ref, pr.state)
    ts = my_slack.chat_post_message(CHANNEL, message).ts

    i = 0

    while pr.state not in ["closed", "merged"]:
        log("polling #%d" % i)
        sleep(5)

        pr = my_github.get_pull_request(
            data.repo.owner.login, data.repo.name, pr.number
        )
        message = "%s meow [%s]" % (pr.links.html.h_ref, pr.state)
        my_slack.chat_update(CHANNEL, ts, message)

        i += 1

        if i % 3 == 0:
            # Spreadsheet contains a list of usernames
            rows = my_googlesheets.read_range(SHEET_ID, "A1:A5")
            the_chosen_one = rows[rand.intn(len(rows))][0]
            log("meowing at %s" % the_chosen_one)
            user_email = "%s@autokitteh.com" % the_chosen_one
            user = my_slack.users_lookup_by_email(user_email).user
            my_slack.chat_post_message(CHANNEL, "paging <@%s>" % user.id, thread_ts=ts)

    log("pr is %s" % pr.state)


def log(msg):
    print("[%s] %s" % (time.now(), msg))
