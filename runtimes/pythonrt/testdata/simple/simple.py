from os import getenv


def greet(event):
    print('INFO: simple: HOME:', getenv('HOME'))
    print('INFO: simple: USER:', getenv('USER'))
    print(f'INFO: simple: event_id: {event["event_id"]}')
    return 'Hello stranger'


if __name__ == '__main__':
    print(greet('garfield'))
