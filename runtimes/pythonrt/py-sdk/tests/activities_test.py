from autokitteh import activities


def test_register_no_activity(monkeypatch):
    no_act = set()
    monkeypatch.setattr(activities, "_no_activity", no_act)

    fns = [dict.get, list.append]
    # Twice to make sure no duplicates
    activities.register_no_activity(fns)
    activities.register_no_activity(fns)

    assert no_act == set(fns)
