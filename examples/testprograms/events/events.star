def on_http_get():
    print("creating subscription")
    postSubscription = subscribe('http_event', 'data.url.path == "/test" && data.method == "POST"')
    getSubscription = subscribe('http_event', 'data.url.path == "/test" && data.method == "GET"')

    print("waiting for event on post subscription")
    print(postSubscription)
    event = next_event(postSubscription)
    print("got event")
    print(event)
    print("waiting for event on post or get")
    print(postSubscription)
    print(getSubscription)
    event = next_event(getSubscription, postSubscription)
    print("got event")
    print(event)

    unsubscribe(getSubscription)
    unsubscribe(postSubscription)
    print("done")

    
