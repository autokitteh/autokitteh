from os import getenv


def greet(event):
    print('INFO: simple: HOME:', getenv('HOME'))
    print('INFO: simple: USER:', getenv('USER'))
    print(f'INFO: simple: event: {event!r}')

    body = event.get('data', {}).get('body')
    if body:
        print(f'BODY: {body!r}')
    return 'Hello stranger'


if __name__ == '__main__':
    print(greet('garfield'))
