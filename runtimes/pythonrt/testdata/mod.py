# Test file for loader, see ../ak_runner_test.py::test_load_code


from datetime import datetime

# Internal function, shouldn't be patched
def log(msg):
    print(msg)


def parse(data):
    # External function, should be patched
    now = datetime.now()
    log(f'{data!r} at {now}')
    return 7
