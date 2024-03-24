import issues
from unittest.mock import MagicMock


def test_on_issue(monkeypatch):
    mock = MagicMock()
    monkeypatch.setattr(issues, 'WebClient', lambda token: mock)

    event = {
        'data': {
            'action': 'opened',
            'issue': {
                'title': 'Fix url',
                'number': 1,
                'user': {'login': 'tebeka'},
                'html_url': 'https://api.github.com/repos/tebeka/toggl/issues/1',
            },
        },
    }

    issues.on_issue(event)

    text = issues.format_message(event['data']['issue'])
    mock.chat_postMessage.assert_called_once_with(channel=issues.CHANNEL_ID, text=text)
