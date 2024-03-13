load("@http", "http1")
load("env", "MEOW")

print(MEOW)

def foo(x):
    return "meow, " + x

def on_http_get():
    resp, err = http1.get("https://httpbin3212.org/status/404", ak.callopts(catch=True))

    # also works:
    # resp, err = catch(lambda: http1.get("https://httpbin3212.org/status/404"))
    # resp, err = catch(http1.get, "https://httpbin3212.org/status/404")
    # resp, err = http1.get("https://httpbin3212.org/status/404", ak = ak.callopts(catch=True))
    # resp, err = http1.get("https://httpbin3212.org/status/404", ak_catch=True)
    # resp, err = http1.get("https://httpbin3212.org/status/404", ak = {"catch": True})

    print(resp, err)
    print(err.op)


def on_http_post(data):
    def again(x):
        return x < 50

    n = poll(lambda: rand.intn(100), again)

    print("zzzz1", n, time.now())
    sleep(1)
    n += 1000
    print("zzzz2", n, time.now())
    sleep(2)
    n += 1000
    print("zzzz3", n, time.now())
    sleep(3)
    resp = http1.get("http://example.com")
    print(resp)
