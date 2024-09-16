"""Markdown-related helper functions across GitHub and Slack."""

load("user_helpers.star", "resolve_github_user", "resolve_slack_user")

def github_markdown_to_slack(text, pr_url, github_owner_org):
    """Convert GitHub markdown text to Slack markdown text.

    References:
    - https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax
    - https://api.slack.com/reference/surfaces/formatting
    - https://github.com/qri-io/starlib/tree/master/re

    Args:
        text: Text body, possibly containing GitHub markdown.
        pr_url: URL of the PR we're working on, used to convert
            other PR references in the text ("#123") to links.
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        Slack markdown text.
    """

    # Split into lines (Qri's "re" module doesn't support the MULTILINE flag).
    lines = text.replace("\r", "").split("\n")

    # Header lines --> bold lines.
    lines = [re.sub(r"^#+\s+(.+)", "**$1**", line) for line in lines]

    # Lists: "-" --> "•" and "◦".
    lines = [re.sub(r"^- ", "  •  ", line) for line in lines]
    lines = [re.sub(r"^  - ", r"          ◦   ", line) for line in lines]
    text = "\n".join(lines)

    # Links: "[text](url)" --> "<url|text>".
    # Images: "![text](url)" --> "Image: <url|text>".
    text = re.sub(r"\[(.*?)\]\((.*?)\)", "<$2|$1>", text)
    text = re.sub(r"!<(.*?)>", "Image: <$1>", text)

    # "@..." --> "<@U...>" or "<https://github.com/...|...>".
    for github_user in re.findall(r"@[\w-]+", text):
        profile_link = pr_url.split(github_owner_org)[0] + github_user[1:]
        user_obj = struct(login = github_user[1:], htmlurl = profile_link)
        slack_user = resolve_github_user(user_obj, github_owner_org)
        text = text.replace(github_user, slack_user)

    # "#123" --> "<PR URL|#123>".
    url_base = re.sub(r"/pull/\d+$", "/pull", pr_url)
    text = re.sub(r"#(\d+)", "<%s/$1|#$1>" % url_base, text)

    # Bold and nested italic text: "***" --> "**_".
    text = re.sub(r"\*\*\*(.+?)\*\*\*", "**_${1}_**", text)

    # Italic text: "*" --> "_".
    text = re.sub(r"(^|[^*])\*([^*]+?)\*", "${1}_${2}_", text)

    # Bold text: "**" or "__" --> "*".
    text = re.sub(r"\*\*(.+?)\*\*", "*$1*", text)
    text = re.sub(r"__(.+?)__", "*$1*", text)

    return text

def slack_markdown_to_github(text, github_owner_org):
    """Convert Slack markdown text to GitHub markdown text.

    References:
    - https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax
    - https://api.slack.com/reference/surfaces/formatting
    - https://github.com/qri-io/starlib/tree/master/re

    Args:
        text: Text body, possibly containing Slack markdown.
        github_owner_org: Required for GitHub org-specific visibility.

    Returns:
        GitHub markdown text.
    """

    # Bold text: "*" --> "**".
    text = re.sub(r"\*(.+?)\*", "**$1**", text)

    # Links: "<url|text>" --> "[text](url)".
    text = re.sub(r"<(.*?)\|(.*?)>", "[$2]($1)", text)

    # Channels: "<#...|name>" --> "[name](#...)" -->
    # "[#name](https://slack.com/app_redirect?channel=...)"
    # (see https://api.slack.com/reference/deep-linking).
    text = re.sub(r"\[(.*?)\]\(#(.*?)\)", "[#$1](https://slack.com/app_redirect?channel=$2)", text)

    # Users: "<@U...>" --> "@github-user".
    for slack_user in re.findall(r"<@[UW][0-9A-Z]*?>", text):
        github_user = resolve_slack_user(slack_user[2:-1], github_owner_org)
        text = text.replace(slack_user, github_user)

    # Multiline code blocks: ```aaa\nbbb``` --> ```\naaa\nbbb\n```.
    text = text.replace("```", "\n```\n")

    # Split into lines (Qri's "re" module doesn't support the MULTILINE flag).
    lines = text.replace("\r", "").split("\n")

    # Quoted text: "&gt; aaa\n&gt; bbb\nccc" --> "> aaa\n> bbb\n\nccc".
    lines = [re.sub(r"^&gt;(.*)", ">$1\n", line) for line in lines]
    text = "\n".join(lines)
    text = text.replace("\n\n>", "\n>")

    return text
