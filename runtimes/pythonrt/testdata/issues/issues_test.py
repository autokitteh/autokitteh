import issues
from unittest.mock import MagicMock


def test_on_issue(monkeypatch):
    mock = MagicMock()
    def webclient(**_): return mock
    monkeypatch.setattr(issues, 'WebClient', webclient)

    event = {
        'data': {
            'action': 'opened',
            'issue': {
                'title': 'Fix url',
                'number': 1,
                'user': {'login': 'tebeka'},
                'htmlurl': 'https://api.github.com/repos/tebeka/toggl/issues/1',
            },
        },
    }

    issues.on_issue(event)
    mock.chat_postMessage.assert_called_once()
