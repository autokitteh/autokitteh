load("@redis", "myredis")

def on_http_get(data):
    myredis.set("foo", "bar")
    store.set("baz", "boop")
    print(myredis.get("foo"))
    print(store.get("baz"))
