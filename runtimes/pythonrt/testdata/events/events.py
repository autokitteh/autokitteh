import autokitteh as ak


def on_http_get(event):
    print("creating subscription")
    postSubscription = ak.subscribe(
        "http_event", 'data.url.path == "/test" && data.method == "POST"'
    )
    getSubscription = ak.subscribe(
        "http_event", 'data.url.path == "/test" && data.method == "GET"'
    )

    print("waiting for event on post subscription")
    print(postSubscription)
    event = ak.next_event(postSubscription)
    print("got event")
    print(event)
    print("waiting for event on post or get")
    print(postSubscription)
    print(getSubscription)
    event = ak.next_event(getSubscription)  # , postSubscription)
    print("got event")
    print(event)

    ak.unsubscribe(getSubscription)
    ak.unsubscribe(postSubscription)
    print("done")
