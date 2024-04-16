load("@http", "myhttp")

def foo(x):
    return "meow, " + x

def on_http_get():
    print(time.now())
    print(store.set("x", "meow"))

    resp, err = myhttp.get("https://httpbin.org/status/404", ak.callopts(catch=True))

    # also works:
    # resp, err = catch(lambda: myhttp.get("https://httpbin3212.org/status/404"))
    # resp, err = catch(myhttp.get, "https://httpbin3212.org/status/404")
    # resp, err = myhttp.get("https://httpbin3212.org/status/404", ak = ak.callopts(catch=True))
    # resp, err = myhttp.get("https://httpbin3212.org/status/404", ak_catch=True)
    # resp, err = myhttp.get("https://httpbin3212.org/status/404", ak = {"catch": True})

    print(resp, err)
    if err:
        print(err.op)
    else:
        print(resp.body.text())

def on_http_post(data):
    addr = data.body.text()
    if not addr:
        addr = "main"

    def again(x):
        return x < 50

    n = poll(lambda: rand.intn(100), again)

    print("zzzz1", n, time.now())
    sleep(1)
    n += 1000

    print(catch(lambda: myhttp.post('http://localhost:9980/webtools/api/msgs/' + addr, raw_body="woof")))

    print("zzzz2", n, time.now())
    sleep(2)

    n += 1000
    print("zzzz3", n, time.now())
    sleep(3)

    print(catch(lambda: myhttp.post('http://localhost:9980/webtools/api/msgs/' + addr, raw_body="meow")))

def on_http_test(data, event, trigger):
    print(data)
    print(event)
    print(trigger)
