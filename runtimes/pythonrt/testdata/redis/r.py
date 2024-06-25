import autokitteh
import redis

r = redis.Redis(host="localhost", port=6379, decode_responses=True)


@autokitteh.activity
def on_event(event):
    print("EVENT:", event)
