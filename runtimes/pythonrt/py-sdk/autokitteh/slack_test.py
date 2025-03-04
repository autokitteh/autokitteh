"""Unit tests for the "slack" module."""

from autokitteh import slack


def test_normalize_channel_name():
    assert slack.normalize_channel_name("") == ""
    assert slack.normalize_channel_name('"isn\'t"') == "isnt"
    assert slack.normalize_channel_name("TEST") == "test"
    assert slack.normalize_channel_name("1  2--3__4") == "1-2-3-4"
    assert slack.normalize_channel_name("a `~!@#$%^&*() 1") == "a-1"
    assert slack.normalize_channel_name("b -_=+ []{}|\\ 2") == "b-2"
    assert slack.normalize_channel_name("c ;:'\" ,.<>/? 3") == "c-3"
    assert slack.normalize_channel_name("-foo ") == "foo"
