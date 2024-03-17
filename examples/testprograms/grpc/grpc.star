load("@grpc", "grpc1")

def on_http_get():
    evs = grpc1.call("localhost:9980", "services.Events1", "List", {})
    print(evs)
