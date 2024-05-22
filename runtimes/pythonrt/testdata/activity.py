import ak
from os import getenv


def handler(message):
    et = phone_home()
    print(et + ' ' + message)

@ak.activity
def phone_home():
    # getenv should not become an activity
    home = getenv('HOME', default='Chicago')
    return f'Gertie @ {home}'
