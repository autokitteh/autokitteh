import os


def greet(name):
    # home = os.getenv('HOME')
    print('INFO: HOME:', os.getenv('HOME')) # home)
    print(f'INFO: greeting {name!r}')
    return 'Hello ' + str(name)


if __name__ == '__main__':
    greet('garfield')
