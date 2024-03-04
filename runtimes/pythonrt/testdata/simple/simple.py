import os


def greet(name):
    print('INFO: simple: HOME:', os.getenv('HOME'))
    print(f'INFO: simple: greeting {name!r}')
    return 'Hello ' + str(name)


if __name__ == '__main__':
    print(greet('garfield'))
