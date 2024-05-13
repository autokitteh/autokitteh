import messages

import pytest

def test_round_trip():
    msg = messages.CallbackMessage(
        name='garfield',
        args=['pizza'],
        kw={'action': 'eat'},
        data=b'odie',
    )
    data = messages.encode_message(msg)
    msg2 = messages.decode_message(data)
    assert msg == msg2

def test_unknown():
    data = '{"type": "no-such-message"}'
    with pytest.raises(ValueError):
        messages.decode_message(data)
