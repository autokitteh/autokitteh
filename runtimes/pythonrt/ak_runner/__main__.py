import ast
import json
import logging
import sys
import tarfile
import traceback
from base64 import b64decode
from os import chdir, mkdir
from pathlib import Path
from socket import AF_UNIX, SOCK_STREAM, socket

from ak_runner import (
    ACTION_NAME,
    AKCall,
    AttrDict,
    Comm,
    is_marked_activity,
    load_code,
    log,
)


def parse_path(root_path):
    """
    >>> parse_path('review.py:on_github_pull_request')
    ('review', 'on_github_pull_request')
    """
    if ":" not in root_path:
        raise ValueError(f"{root_path!r} - missing :")

    file_name, func_name = root_path.split(":", 1)
    if not file_name.endswith(".py"):
        raise ValueError(f"{root_path!r} - not a Python file")

    return file_name[:-3], func_name


def extract_code(tar_path):
    root_dir = Path(tar_path).absolute().parent
    code_dir = root_dir / "code"
    mkdir(code_dir)
    with tarfile.open(tar_path) as tf:
        tf.extractall(code_dir, filter="data")

    return code_dir


def module_entries(mod):
    return [
        name
        for name in dir(mod)
        if name != ACTION_NAME and callable(getattr(mod, name, None))
    ]


class ActivityFn:
    """Run top level functions as activities."""

    # We can't use lambdas to wrap the original function since it can't be pickled.
    def __init__(self, ak_call, fn):
        self.ak_call = ak_call
        self.fn = fn

    def __call__(self, *args, **kw):
        return self.ak_call(self.fn, *args, **kw)


def run(args):
    sock = socket(AF_UNIX, SOCK_STREAM)
    sock.connect(args.sock)
    comm = Comm(sock)
    log.init(logging.INFO, comm)
    log.info("connected to %r", args.sock)

    module_name = args.path[:-3]  # Trim .py suffix

    py_version = "{}.{}".format(*sys.version_info[:2])
    log.info("python: %r, version: %r", sys.executable, py_version)
    log.info("sock: %r, tar: %r, module: %r", args.sock, args.tar, module_name)
    code_dir = extract_code(args.tar)
    log.info("code dir: %r", code_dir)

    py_file = code_dir / args.path
    if not py_file.exists():
        raise SystemExit(f"error: {py_file.name!r} not found")

    # Allow users to import their own files and load data files
    sys.path.append(str(code_dir))
    chdir(code_dir)

    log.info("loading %r", module_name)
    ak_call = AKCall(comm)
    mod = load_code(code_dir, ak_call, module_name)
    ak_call.set_module(mod)

    entries = module_entries(mod)
    comm.send_exported(entries)

    # Initial call.
    message = comm.receive_run()
    func_name = message.get("func_name")
    if func_name is None:
        log.error("no function name in %r", message)
        raise SystemExit(1)

    fn = getattr(mod, func_name, None)
    if fn is None:
        log.error("%r has no function %r", module_name, func_name)
        raise SystemExit(1)

    # Support activity decorator in top level handlers
    if is_marked_activity(fn):
        fn = ActivityFn(ak_call, fn)

    event = message.get("event")
    event = {} if event is None else event

    # Inject HTTP body
    # TODO (ENG-624) change this once we support callbacks to autokitteh
    body = event.get("data", {}).get("body", {}).get("bytes")
    if isinstance(body, str):
        try:
            event["data"]["body"]["bytes"] = b64decode(body)
        except ValueError:
            pass

    event = AttrDict(event)
    try:
        fn(event)
    except Exception as err:
        log.exception("error running %s: %s", func_name, err)
        # Print the error to stderr so it'll show in session logs
        print(f"error: {err}", file=sys.stderr)
        traceback.print_exception(err, file=sys.stderr)
        comm.send_error(err)
        raise SystemExit(1)
    comm.send_done()


def inspect_file(code_dir, path):
    with open(path) as fp:
        code = fp.read()

    tree = ast.parse(code, path.name, "exec")
    for node in tree.body:
        if not isinstance(node, (ast.FunctionDef, ast.ClassDef)):
            continue

        export = {
            "name": node.name,
            "file": str(path.relative_to(code_dir)),
            "line": node.lineno,
        }
        yield export


def inspect(args):
    code_dir = Path(args.path)

    entries = []
    for path in code_dir.glob("**/*.py"):
        try:
            entries.extend(inspect_file(code_dir, path))
        except Exception as err:
            raise RuntimeError(f"inspecting {path}: {err}")

    # Stdout is read by Go, don't print anything else
    print(json.dumps(entries), file=args.output)


# argparse.FileType will open the file
def file_type(value):
    path = Path(value)
    if path.is_file() or path.is_socket():
        return value

    raise ValueError(f"{value!r} - not a file")


def dir_type(value):
    path = Path(value)
    if path.is_dir():
        return value

    raise ValueError(f"{value!r} - not a directory")


if __name__ == "__main__":
    import sys
    from argparse import ArgumentParser, FileType

    parser = ArgumentParser(prog="ak_runner", description="autokitteh Python runner")
    sp = parser.add_subparsers(help="sub command help", required=True)

    parse_run = sp.add_parser("run", help="run user code")
    parse_run.add_argument("sock", help="path to unix domain socket", type=file_type)
    parse_run.add_argument("tar", help="path to code tar file", type=file_type)
    parse_run.add_argument("path", help="file.py:function")
    parse_run.set_defaults(func=run)

    parse_inspect = sp.add_parser("inspect", help="inspect user code")
    parse_inspect.add_argument("path", help="path to code", type=dir_type)
    parse_inspect.add_argument("output", help="output file", type=FileType("w"))
    parse_inspect.set_defaults(func=inspect)

    args = parser.parse_args()
    args.func(args)
