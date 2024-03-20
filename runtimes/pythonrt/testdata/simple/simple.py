import os


def greet(event):
    print('INFO: simple: HOME:', os.getenv('HOME'))
    print(f'INFO: simple: event_id: {event["event_id"]}')
    return 'Hello stranger'


if __name__ == '__main__':
    print(greet('garfield'))
