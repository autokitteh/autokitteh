import ak


def handler(message):
    et = phone_home()
    print(et + ' ' + message)

@ak.activity
def phone_home():
    return 'Gertie'
