import ast
import builtins
import sys
from importlib.util import spec_from_file_location
from pathlib import Path

import log


def name_of(node, code_lines):
    line = code_lines[node.lineno - 1]
    name = line[node.col_offset : node.end_col_offset]
    return name


AK_CALL_NAME = "_ak_call"
BUILTIN = {v for v in dir(builtins) if callable(getattr(builtins, v))}


class Transformer(ast.NodeTransformer):
    """Replace 'fn(a, b)' with '_ak_call(fn, a, b)'."""

    def __init__(self, file_name, src):
        self.file_name = file_name
        self.code_lines = src.splitlines()
        # Indent of class/function definition. -1 means top level code
        self.fn_indent = -1

    def visit(self, node):
        # Visit AST nodes. We keep track of functions and indent, in order not to patch
        # module level calls.
        indent = getattr(node, "col_offset", None)
        if indent is None:  # Module and others don't have col_offset
            self.generic_visit(node)
            return node

        if isinstance(node, (ast.FunctionDef, ast.ClassDef)) and self.fn_indent == -1:
            self.fn_indent = indent
            self.generic_visit(node)
            return node

        if indent <= self.fn_indent:
            self.fn_indent = -1

        if not isinstance(node, ast.Call) or self.fn_indent == -1:
            self.generic_visit(node)
            return node

        name = name_of(node.func, self.code_lines)

        if not name or name in BUILTIN:
            self.generic_visit(node)
            return node

        log.info("%s:%d: patching %s with ak_call", self.file_name, node.lineno, name)
        # urlopen("https://autokitteh.h") -> _call(urlopen, "https://autokitteh.com")
        call = ast.Call(
            func=ast.Name(id=AK_CALL_NAME, ctx=ast.Load()),
            args=[node.func] + node.args,
            keywords=node.keywords,
        )
        self.generic_visit(node)
        return call


class Loader:
    """Implement importlib.Loader"""

    def __init__(self, ak_call):
        self.ak_call = ak_call

    def exec_module(self, module):
        log.info("importing %r", module.__file__)

        with open(module.__file__) as fp:
            src = fp.read()

        tree = ast.parse(src, module.__file__, "exec")
        trans = Transformer(module.__file__, src)
        patched_tree = trans.visit(tree)
        ast.fix_missing_locations(patched_tree)

        code = compile(patched_tree, module.__file__, "exec")
        setattr(module, AK_CALL_NAME, self.ak_call)
        exec(code, module.__dict__)

    def create_module(self, spec):
        return None  # Use default module creation


class Finder:
    """An importlib finder that will handler files from user code directory."""

    def __init__(self, code_dir, ak_call):
        self.code_dir = code_dir
        self.ak_call = ak_call

    def find_spec(self, fullname: str, path: list[str], target=None):
        if path:
            mod_path = Path(path[0])
            if not mod_path.is_relative_to(self.code_dir):
                return None

        relative_path = fullname.replace(".", "/")  # json.decoder -> json/decoder
        # NOTE: We currently don't support packages (directory with __init__.py)
        # We'll consider that once there's a concrete user request
        full_path = self.code_dir / (relative_path + ".py")
        if not full_path.is_file():
            return None

        loader = Loader(self.ak_call)
        spec = spec_from_file_location(fullname, full_path, loader=loader)
        return spec


def load_code(root_path: Path, ak_call, module_name: str):
    """Load user code, patch function calls."""
    finder = Finder(root_path, ak_call)
    try:
        sys.meta_path.insert(0, finder)
        return __import__(module_name)
    finally:
        sys.meta_path.pop(0)


def fn_args(node):
    """Return list of arguments to fn (AST)."""
    args = [a.arg for a in node.args.args]
    if node.args.kwarg:
        args.append(node.args.kwarg.arg)

    return args


def class_args(node):
    for n in node.body:
        if not isinstance(n, ast.FunctionDef):
            continue
        if n.name != "__init__":
            continue

        args = fn_args(n)
        return args[1:]  # Remove self


def exports(code_dir, file_name):
    """Returns an iterator of functions & classes defined in file_name."""
    full_path = code_dir / file_name
    with open(full_path) as fp:
        code = fp.read()

    tree = ast.parse(code, file_name, "exec")
    for node in tree.body:
        if not isinstance(node, (ast.FunctionDef, ast.ClassDef)):
            continue

        args = fn_args(node) if isinstance(node, ast.FunctionDef) else class_args(node)
        yield {
            "file": str(file_name),
            "line": node.lineno,
            "name": node.name,
            "args": args,
        }
