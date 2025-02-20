from random import random

from .decorators import activity, noactivity


def test_default(*args, **kwargs):
    return kwargs.get("ret")


@activity
def test_activity(*args, **kwargs):
    return kwargs.get("ret")


@noactivity
def test_noactivity(*args, **kwargs):
    return kwargs.get("ret")


@noactivity
def test_noactivity_internal_call(*args, **kwargs):
    return test_default(*args, **kwargs)


@noactivity
def test_noactivity_external_call(*args, **kwargs):
    _ = random()
    return kwargs.get("ret")
