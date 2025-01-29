import re
from base64 import b64decode
from datetime import datetime

import pytest

import deterministic

nonact_caes = [
    # function, result
    (b64decode, True),
    (datetime.now, False),
    (datetime.strptime, True),
    ([].append, True),
    (re.search("[a-z]", "hello").group, True),
]


@pytest.mark.parametrize("func, expected", nonact_caes)
def test_is_deterministic(func, expected):
    out = deterministic.is_deterministic(func)
    assert out == expected, func.__qualname__
