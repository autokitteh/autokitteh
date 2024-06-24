import autokitteh
import pytest


def test_AttrDict():
    cfg = autokitteh.AttrDict({
        'server': {
            'port': 8080,
            'interface': 'localhost',
        },
        'mode': 'dev',
        'logging': {
            'level': 'info',
        },
    })
    assert cfg['server']['port'] == cfg.server.port
    assert cfg['mode'] == cfg.mode

    with pytest.raises(NotImplementedError):
        cfg.server.port = 8081

