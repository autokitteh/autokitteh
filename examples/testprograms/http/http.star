load("@http", "http1")
load("env", "MEOW")

print(MEOW)

def on_http_get():
    resp = http1.get("https://httpbin3212.org/status/404")
    if resp.status_code != 200:
        print("oh noes")
    print(resp)

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
