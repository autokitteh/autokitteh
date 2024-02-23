def on_http_get():
    print("creating subscription")
    postSubscription = subscribe('http_event', "post")
    getSubscription = subscribe('http_event', "get")

    print("waiting for event on post subscription")
    print(postSubscription)
    event = next_event(postSubscription)
    print("got event")
    print(event)
    print("waiting for event on post or get")
    print(postSubscription)
    print(getSubscription)
    event = next_event(postSubscription, getSubscription)
    print("got event")
    print(event)

    unsubscribe(getSubscription)
    unsubscribe(postSubscription)
    print("done")

    
