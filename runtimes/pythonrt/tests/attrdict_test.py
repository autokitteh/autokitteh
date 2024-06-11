import pytest

from ak_runner.attrdict import AttrDict


def test_AttrDict():
    cfg = AttrDict({
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
