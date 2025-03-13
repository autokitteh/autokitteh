import autokitteh


def on_event(event):
    print("START:", event)
    other(event)
    print("END")


@autokitteh.activity
def other(event):
    print("OTHER")
