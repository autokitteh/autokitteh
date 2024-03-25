def on_http_get_meow(data, trigger, event):
    print(data)
    print(trigger)
    print(event)
    print(http.get("http://example.com/{}".format(trigger.data.params['who'])))
