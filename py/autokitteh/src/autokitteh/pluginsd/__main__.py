import sys

from .pluginsd import serve

if __name__ == '__main__':
    sys.path.insert(0, '.')
    serve()
