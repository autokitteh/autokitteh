import re
from base64 import b64decode
from datetime import datetime

import pytest

import deterministic

nonact_case = [
    # function, result
    (b64decode, True),
    (datetime.now, False),
    (datetime.strptime, True),
    ([].append, True),
    (re.search("[a-z]", "hello").group, True),
]


@pytest.mark.parametrize("func, expected", nonact_case)
def test_is_deterministic(func, expected):
    out = deterministic.is_deterministic(func)
    assert out == expected, func.__qualname__


class NoAct:
    def method(self):
        pass

    @staticmethod
    def static():
        pass

    @classmethod
    def klass(cls):
        pass


def no_act_fn():
    pass


def yes_act_fn():
    pass


no_acts = {NoAct.method, NoAct.static, NoAct.klass, no_act_fn, Exception.__init__}
no_act_cases = [
    (no_act_fn, True),
    (NoAct.method, True),
    (NoAct.static, True),
    (NoAct.klass, True),
    (yes_act_fn, False),
    (Exception.__init__, True),
]


@pytest.mark.parametrize("func, expected", no_act_cases)
def test_is_no_activity(func, expected, monkeypatch):
    monkeypatch.setattr(deterministic.activities, "_no_activity", no_acts)
    out = deterministic.is_no_activity(func)
    assert out == expected, func.__qualname__


nact = NoAct()
no_act_cls_cases = [
    [nact.method, True],
    [nact.static, True],
    [nact.klass, True],
    [NoAct.static, True],
    [NoAct.klass, True],
    (Exception.__init__, False),
]


@pytest.mark.parametrize("func, expected", no_act_cls_cases)
def test_is_no_activity_class(func, expected, monkeypatch):
    monkeypatch.setattr(deterministic.activities, "_no_activity", {NoAct})
    out = deterministic.is_no_activity(func)
    assert out == expected, func.__qualname__
