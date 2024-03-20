load("@grpc", "grpc1")

def on_http_get():
    evs2 = grpc1.call(host="localhost:9980", service="autokitteh.events.v1.EventsService", method="List", payload={"event_type":"get"})
    print(evs2)
