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
from importlib.util import spec_from_file_location, module_from_spec
from os import mkdir
from pathlib import Path
from socket import AF_UNIX, SOCK_STREAM, socket
from time import sleep
from types import ModuleType

import autokitteh

log: logging.Logger = None  # Filled in main

def name_of(node):
    """Name of call node (e.g. 'requests.get')"""
    if isinstance(node, ast.Attribute):
        prefix = name_of(node.value)
        return f'{prefix}.{node.attr}'

    if isinstance(node, ast.Call):
        return name_of(node.func)

    if isinstance(node, ast.Name):
        return node.id

    raise ValueError(f'unknown AST node type: {node!r}')


ACTION_NAME = '_ak_call'
BUILTIN = {v for v in dir(builtins) if callable(getattr(builtins, v))}


class Transformer(ast.NodeTransformer):
    """Replace 'fn(a, b)' with '_ak_call(fn, a, b)'."""
    def __init__(self, file_name):
        self.file_name = file_name

    def visit_Call(self, node):
        # Recurse, see https://docs.python.org/3/library/ast.html#ast.NodeVisitor.generic_visit
        self.generic_visit(node)

        name = name_of(node.func)

        if not name or name in BUILTIN:
            return node

        log.info('%s:%d: patching %s with action', self.file_name, node.lineno, name)
        call = ast.Call(
            func=ast.Name(id=ACTION_NAME, ctx=ast.Load()),
            args=[node.func] + node.args,
            keywords=node.keywords,
        )
        return call


def load_code(root_path, action_fn, module_name):
    """Load user code into a module, instrumenting function calls."""
    log.info('importing %r', module_name)
    file_name = Path(root_path) / (module_name + '.py')
    with open(file_name) as fp:
        src = fp.read()

    tree = ast.parse(src, file_name, 'exec')
    trans = Transformer(file_name)
    patched_tree = trans.visit(tree)
    ast.fix_missing_locations(patched_tree)

    code = compile(patched_tree, file_name, 'exec')

    module = ModuleType(module_name)
    setattr(module, ACTION_NAME, action_fn)
    exec(code, module.__dict__)

    return module


def run_code(mod, entry_point, data):
    fn = getattr(mod, entry_point, None)
    if fn is None:
        raise NameError('%s.%s not found', mod.__name__, entry_point)

    log.info('calling %s.%s', mod.__name__, entry_point)
    return fn(data)


class MessageType:
    callback = 'callback'
    done = 'done'
    log = 'log'
    module = 'module'
    response = 'response'
    run = 'run'
    call = 'call'
    call_return = 'return'
    

class Comm:
    def __init__(self, sock):
        self.sock = sock
        self.rdr = sock.makefile('r')

    def _send(self, message):
        data = json.dumps(message) + '\n'
        self.sock.sendall(data.encode('utf-8'))

    def recv(self, *msg_types):
        data = self.rdr.readline()
        if not data:
            raise ValueError('connection closed')

        message = json.loads(data)
        if (typ := message['type']) not in msg_types:
            typs = ', '.join(msg_types)
            raise ValueError(f'message type: expected one of {typs!r}, got {typ!r}')
        return message

    def _picklize(self, data):
        data = pickle.dumps(data, protocol=0)
        return b64encode(data).decode('utf-8')

    def send_activity(self, fn, args, kw):
        data = (fn, args, kw)
        message = {
            'type': MessageType.callback,
            'payload': {
                'name': fn if isinstance(fn, str) else fn.__name__,
                'args': [repr(a) for a in args],
                'kw': {k: repr(v) for k, v in kw.items()},
                'data': self._picklize(data),
            },
        }
        self._send(message)

    def extract_activity(self, message):
        payload = message['payload']
        data = b64decode(payload['data'])
        payload['data'] = pickle.loads(data)
        return payload

    def send_exported(self, entries):
        message = {
            'type': MessageType.module,
            'payload': {
                'entries': entries,
            }
        }
        self._send(message)

    def send_done(self):
        message = {'type': MessageType.done}
        self._send(message)

    def receive_run(self):
        message = self.recv(MessageType.run)
        return message['payload']

    def send_response(self, value):
        message = {
            'type': MessageType.response,
            'payload': {
                'value': self._picklize(value),
            }
        }
        self._send(message)

    def extract_response(self, message):
        data = message['payload']['value']
        return pickle.loads(b64decode(data))


    def send_log(self, level, message):
        message = {
            'type': MessageType.log,
            'payload': {
                'level': level,
                'message': message,
            },
        }

    def send_call(self, func_name, args):
        message = {
            'type': MessageType.call,
            'payload': {
                'func_name': func_name,
                'args': args,
            },
        }
        self._send(message)

# Functions that are called back to ak
AK_FUNCS = {
    autokitteh.next_event,
    autokitteh.subscribe,
    autokitteh.unsubscribe,

    sleep,
}


class AKCall:
    """Callable wrapping functions with activities."""
    def __init__(self, comm: Comm):
        self.in_activity = False
        self.comm = comm
        self.module = None

    def is_module_func(self, fn):
        return fn.__module__ == self.module.__name__

    def should_run_as_activity(self, fn):
        if self.in_activity:
            return False

        if getattr(fn, autokitteh.ACTIVITY_ATTR, False):
            return True

        if fn.__module__ == 'builtins':
            return False
        
        if self.is_module_func(fn):
            return False

        return True

    def __call__(self, func, *args, **kw):
        if not self.should_run_as_activity(func):
            log.info(
                'calling %s (args=%r, kw=%r) directly (in_activity=%s)', 
                func.__name__, args, kw, self.in_activity)
            return func(*args, **kw)

        log.info('ACTION: calling %s via activity (args=%r, kw=%r)', func.__name__, args, kw)
        self.in_activity = True
        try:
            if func in AK_FUNCS:
                self.comm.send_call(func.__name__, args)
                msg = self.comm.recv(MessageType.call_return)
                value = msg['payload']['value']
                if func is autokitteh.next_event:
                    value = autokitteh.AttrDict(value)
                return value

            if self.is_module_func(func):
                # Pickle can't handle function from our loaded module
                func = func.__name__
            self.comm.send_activity(func, args, kw)
            message = self.comm.recv(MessageType.callback, MessageType.response)
            
            if message['type'] == MessageType.callback:
                payload = self.comm.extract_activity(message)
                fn, args, kw = payload['data']
                if isinstance(fn, str):
                    fn = getattr(self.module, fn, None)
                    if fn is None:
                        mod_name = self.module.__name__
                        raise ValueError(f'function {fn} not found in {mod_name}')
                value = fn(*args, **kw)
                self.comm.send_response(value)
                message = self.comm.recv(MessageType.response)

            # Reply message, either from current call or playback
            return self.comm.extract_response(message)
        finally:
            self.in_activity = False


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


def encode_msg(typ, name, payload, func_name="", func_args=None, kw=None):
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
            'kw': kw or {},
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


class AKLogHandler(logging.Handler):
    def __init__(self, level, comm):
        super().__init__(level)
        self.comm = comm
        self.formatter = logging.Formatter()

    def emit(self, record):
        level = 'ERROR' if record.levelname == 'CRITICAL' else record.levelname
        message = record.getMessage()
        if record.exc_info:
            message += '\n' + self.formatter.formatException(record.exc_info)
        self.comm.send_log(level, message)

def create_logger(level, comm):
    log = logging.getLogger('AK')
    log.setLevel(level)
    handler = AKLogHandler(level, comm)
    log.addHandler(handler)
    return log



def run(args):
    global log 

    sock = socket(AF_UNIX, SOCK_STREAM)
    sock.connect(args.sock)
    comm = Comm(sock)
    log = create_logger(logging.INFO, comm)

    # TODO: Ask Itay why AK does not pass entry point
    if ':' in args.path:
        module_name, _ = parse_path(args.path)
    else:
        module_name = args.path[:-3]

    py_version = '{}.{}'.format(*sys.version_info[:2])
    log.info('python: %r, version: %r', sys.executable, py_version)
    log.info('sock: %r, tar: %r, module: %r', args.sock, args.tar, module_name)
    code_dir = extract_code(args.tar)
    log.info('code dir: %r', code_dir)

    log.info('connected to %r', args.sock)

    log.info('loading %r', module_name)

    ak_call = AKCall(comm)
    mod = load_code(code_dir, ak_call, module_name)
    ak_call.module = mod

    entries = module_entries(mod)
    comm.send_exported(entries)

    # Initial call.
    message = comm.receive_run()
    func_name = message.get('func_name')
    if func_name is None:
        log.error('no function name in %r', message)
        raise SystemExit(1)

    fn = getattr(mod, func_name, None)
    if fn is None:
        log.error('%r has no function %r', module_name, func_name)
        raise SystemExit(1)
        
    event = message.get('event')
    event = {} if event is None else event

    # Inject HTTP body
    # TODO (ENG-624) change this once we support callbacks to autokitteh
    body = event.get('data', {}).get('body')
    if isinstance(body, str):
        try:
            event['data']['body'] = b64decode(body)
        except ValueError:
            pass

    event = autokitteh.AttrDict(event)
    try:
        fn(event)
    except Exception as err:
        log.exception('error running %s: %s', func_name, err)
        raise SystemExit(1)
    comm.send_done()


def inspect_file(root_dir, path):
    mod_name = path.stem
    spec = spec_from_file_location(mod_name, path)
    if spec is None:
        raise ValueError('no spec for {path!r}')

    mod = module_from_spec(spec)
    spec.loader.exec_module(mod)

    for name in dir(mod):
        value = getattr(mod, name)
        if not callable(value):
            continue

        if value.__module__ != mod.__name__:
            continue

        export = {
            'name': name,
            'file': str(path.relative_to(root_dir)),
            'line': value.__code__.co_firstlineno,
        }
        yield export
        

def inspect(args):
    ak_name = Path(__file__).name

    code_dir = Path(args.path)
    entries = []
    for path in code_dir.glob('**/*.py'):
        if path.name == ak_name:
            continue
        entries.extend(inspect_file(code_dir, path))

    # Stdout is read by Go, don't print anything else
    print(json.dumps(entries))


# argparse.FileType will open the file
def file_type(value):
    path = Path(value)
    if path.is_file() or path.is_socket():
        return value

    raise ValueError(f'{value!r} - not a file')


def dir_type(value):
    path = Path(value)
    if path.is_dir():
        return value

    raise ValueError(f'{value!r} - not a directory')


if __name__ == '__main__':
    import sys
    from argparse import ArgumentParser

    parser = ArgumentParser(description='autokitteh Python runner')
    sp = parser.add_subparsers(help='sub command help')

    parse_run = sp.add_parser('run', help='run user code')
    parse_run.add_argument('sock', help='path to unix domain socket', type=file_type)
    parse_run.add_argument('tar', help='path to code tar file', type=file_type)
    parse_run.add_argument('path', help='file.py:function')
    parse_run.set_defaults(func=run)

    parse_inspect = sp.add_parser('inspect', help='inspect user code')
    parse_inspect.add_argument('path', help='path to code', type=dir_type)
    parse_inspect.set_defaults(func=inspect)

    args = parser.parse_args()
    args.func(args)
