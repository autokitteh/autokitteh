"""Run user code under AutoKitteh"""

# This file is long, but keeping it a single file helps with embedding it in Go and
# running it.

import ast
import builtins
import json
import logging
import pickle
import sys
import tarfile
from base64 import b64decode, b64encode
from collections import namedtuple
from functools import wraps
from importlib.abc import Loader
from importlib.machinery import SourceFileLoader
from inspect import isbuiltin
from os import mkdir
from pathlib import Path
from queue import Queue
from socket import AF_UNIX, SOCK_STREAM, socket
from threading import Thread

# TODO: Log to AutoKitteh (ENG-552)
logging.basicConfig(
    format='%(asctime)s - %(levelname)s - %(filename)s:%(lineno)d - %(message)s',
    datefmt='%Y-%M-%DT%H:%M:%S',
    level=logging.INFO,
)


def name_of(node):
    """Name of call node (e.g. 'requests.get')"""
    if isinstance(node, ast.Name):
        return node.id

    prefix = name_of(node.value)
    return f'{prefix}.{node.attr}'


ACTION_NAME = '_ak_call'
MODULE_NAME = ''
BUILTIN = {v for v in dir(builtins) if callable(getattr(builtins, v))}


class Transformer(ast.NodeTransformer):
    """Replace 'fn(a, b)' with '_ak_call(fn, a, b)'"""
    def visit_Call(self, node):
        name = name_of(node.func)
        # ast.Transformer does not recurse to args
        node.args = [self.visit(a) for a in node.args]

        if not name or name in BUILTIN:
            return node

        logging.info('patching %s with action', name)
        call = ast.Call(
            func=ast.Name(id=ACTION_NAME, ctx=ast.Load()),
            args=[node.func] + node.args,
            keywords=node.keywords,
        )
        return call


class AKLoader(Loader):
    """Custom file loaders that will rewrite function calls to actions."""
    def __init__(self, src_loader, action):
        self.file_name = src_loader.path
        self.module_name = src_loader.name
        self.action = action

    def create_module(self, _):
        # Must be defined since it's an abstract method
        return None  # Use default module

    def exec_module(self, module):
        try:
            with open(self.file_name) as fp:
                src = fp.read()
        except OSError as err:
            raise ImportError(f'cannot read {self.module_name!r} - {err}')

        mod = ast.parse(src, self.file_name, 'exec')
        trans = Transformer()
        out = trans.visit(mod)
        ast.fix_missing_locations(out)

        code = compile(out, self.file_name, 'exec')
        setattr(module, ACTION_NAME, self.action)
        exec(code, module.__dict__)


# There is an established way to add import hooks, but we want to change the behavior of
# the current PathFinder found in sys.import_hooks so it'll call our code when importing
# form the user directory. This is why you'll sell all these monkey patches below.

def patch_finder(finder, action):
    """Patches the finder to use a custom source loader."""
    _orig_find_spec = finder.find_spec
    def find_spec(fullname, target=None):
        spec = _orig_find_spec(fullname, target)
        if spec is None or not isinstance(spec.loader, SourceFileLoader):
            return spec

        logging.info('patching loader for %r', fullname)
        spec.loader = AKLoader(spec.loader, action)
        return spec

    finder.find_spec = find_spec


def wrap_hook(hook, user_dir, action):
    """Wraps a hook to patch finder on user code directories."""
    @wraps(hook)
    def wrapper(path):
        finder = hook(path)
        if user_dir.is_relative_to(path):
            patch_finder(finder, action)
        return finder

    return wrapper


def patch_import_hooks(user_dir, action_fn):
    """Patches standard import hook in sys.path_hooks."""
    user_dir = Path(user_dir)
    for i, hook in enumerate(sys.path_hooks):
        if hook.__name__ == 'path_hook_for_FileFinder':
            sys.path_hooks[i] = wrap_hook(hook, user_dir, action_fn)
            return

    raise RuntimeError(f'cannot find import hook to patch in {sys.path_hooks}')


def load_code(root_path, action_fn, module_name):
    patch_import_hooks(root_path, action_fn)
    sys.path.insert(0, str(root_path))
    logging.info('importing %r', module_name)
    mod = __import__(module_name)
    return mod


def run_code(mod, entry_point, data):
    fn = getattr(mod, entry_point, None)
    if fn is None:
        raise NameError('%s.%s not found', mod.__name__, entry_point)

    logging.info('calling %s.%s', mod.__name__, entry_point)
    return fn(data)



# Queue for passing requests from execution thread to main working with Go.
activity_request, activity_response = Queue(), Queue()
Call = namedtuple('Call', 'func args')

class AKCall:
    """Callable wrapping functions with activities."""
    def __init__(self, module_name):
        self.module_name = module_name

    def ignore(self, fn):
        if isbuiltin(fn):
            return True
        
        if fn.__module__ == self.module_name:
            return True

        return False

    def __call__(self, func, *args, **kw):
        if self.ignore(func):
            return func(*args, **kw)

        logging.info('ACTION: calling %s (args=%r, kw=%r)', func.__name__, args, kw)
        request = Call(func, args)
        activity_request.put(request)
        response = activity_response.get()
        return response


class RunWrapper:
    """Wrapper that captures the module so we can access it outside the running
    thread. And also send sentinel to activity_request queue to signal we're done."""
    def __init__(self, mod):
        self.mod = mod

    def run(self, func_name, data):
        run_code(self.mod, func_name, data)
        activity_request.put(None)  # Signal we're done


def extract_code(tar_path):
    root_dir = Path(tar_path).absolute().parent
    code_dir = f'{root_dir}/code'
    mkdir(code_dir)
    with tarfile.open(tar_path) as tf:
        tf.extractall(code_dir)

    return code_dir


def parse_path(root_path):
    """
    >>> parse_path('review.py:on_github_pull_request')
    ('review', 'on_github_pull_request')
    """
    if ':' not in root_path:
        raise ValueError(f'{root_path!r} - missing :')

    file_name, func_name = root_path.split(':', 1)
    if not file_name.endswith('.py'):
        raise ValueError(f'{root_path!r} - not a Python file')

    return file_name[:-3], func_name



# argparse.FileType will open the file
def file_type(value):
    path = Path(value)
    if path.is_file() or path.is_socket():
        return value

    raise ValueError(f'{value!r} - not a file')


def encode_msg(typ, name, payload, func_name="", func_args=None):
    if isinstance(payload, str):
        payload = payload.encode('utf-8')
    data = b64encode(payload)

    data = json.dumps({
        'type': typ,
        'name': name,
        'payload': data.decode('utf-8'),
        'func': {
            'name': func_name,
            'args': func_args or [],
        },
    }) + '\n'
    return data.encode('utf-8')


def decode_msg(data):
    obj = json.loads(data)
    if obj.get('payload'):
        obj['payload'] = b64decode(obj['payload'])

    return obj


def module_entries(mod):
    return [
        name
        for name in dir(mod)
        if name != ACTION_NAME and callable(getattr(mod, name, None))
    ]


if __name__ == '__main__':
    from argparse import ArgumentParser
    import sys

    parser = ArgumentParser(description='autokitteh Python runner')
    parser.add_argument('sock', help='path to unix domain socket', type=file_type)
    parser.add_argument('tar', help='path to code tar file', type=file_type)
    parser.add_argument('path', help='file.py:function')
    args = parser.parse_args()

    # TODO: Ask Itay why AK does not pass entry point
    if ':' in args.path:
        module_name, _ = parse_path(args.path)
    else:
        module_name = args.path[:-3]

    py_version = '{}.{}'.format(*sys.version_info[:2])
    logging.info('python: %r, version: %r', sys.executable, py_version)
    logging.info('sock: %r, tar: %r, module: %r', args.sock, args.tar, module_name)
    code_dir = extract_code(args.tar)
    logging.info('code dir: %r', code_dir)

    sock = socket(AF_UNIX, SOCK_STREAM)
    sock.connect(args.sock)
    rdr = sock.makefile('r')
    logging.info('connected to %r', args.sock)

    logging.info('loading %r', module_name)
    ak_call = AKCall(module_name)
    mod = load_code(code_dir, ak_call, module_name)
    MODULE_NAME = mod.__name__
    entries = module_entries(mod)
    event = encode_msg('module', '', json.dumps(entries))
    sock.sendall(event)

    # Initial call
    request = decode_msg(rdr.readline())
    if request['type'] != 'run':
        logging.error('bad initial request: %r', request)
        raise SystemExit(1)

    func_name = request.get('name')
    if func_name is None:
        logging.error('no function name in %r', request)
        raise SystemExit(1)

    event = request.get('payload')
    event = {} if event is None else json.loads(event)

    rw = RunWrapper(mod)
    Thread(target=rw.run, args=(func_name, event), daemon=True).start()
    logging.info('execution thread started')

    while True:
        request = activity_request.get()
        if request is None:  # Done
            break

        # Use protocol 0 since it's less Python version specific
        event = pickle.dumps(request, protocol=0)
        fn, args = request
        msg = encode_msg('activity', '', event, fn.__name__, [str(a) for a in args])
        logging.info('sending activity request')
        sock.sendall(msg)
        event = rdr.readline()
        logging.info('got activity response')
        resp = decode_msg(event)
        logging.info('activity response: %r', resp)
        fn, args = pickle.loads(resp['payload'])
        logging.info('activity request: %s %r', fn, args)
        out = fn(*args)
        event = pickle.dumps(out, protocol=0)
        msg = encode_msg('response', '', event)
        sock.sendall(msg)
        activity_response.put(out)

    msg = encode_msg('done', '', '')
    sock.sendall(msg)
