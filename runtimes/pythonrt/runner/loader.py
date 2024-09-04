import ast
import builtins
import sys
from pathlib import Path
from types import ModuleType

import log


def name_of(node, code_lines):
    line = code_lines[node.lineno - 1]
    name = line[node.col_offset : node.end_col_offset]
    return name


ACTION_NAME = "_ak_call"
BUILTIN = {v for v in dir(builtins) if callable(getattr(builtins, v))}


class Transformer(ast.NodeTransformer):
    """Replace 'fn(a, b)' with '_ak_call(fn, a, b)'."""

    def __init__(self, file_name, src):
        self.file_name = file_name
        self.code_lines = src.splitlines()

    def visit_Call(self, node):
        # Recurse, see https://docs.python.org/3/library/ast.html#ast.NodeVisitor.generic_visit
        self.generic_visit(node)

        name = name_of(node.func, self.code_lines)

        if not name or name in BUILTIN:
            return node

        log.info("%s:%d: patching %s with action", self.file_name, node.lineno, name)
        call = ast.Call(
            func=ast.Name(id=ACTION_NAME, ctx=ast.Load()),
            args=[node.func] + node.args,
            keywords=node.keywords,
        )
        return call


def load_code(root_path: Path, action_fn, module_name: str):
    """Load user code into a module, instrumenting function calls.

    root_path must be in sys.path if there are local imports in the loaded module.
    """
    log.info("importing %r", module_name)
    file_name = root_path / (module_name + ".py")
    with open(file_name) as fp:
        src = fp.read()

    tree = ast.parse(src, file_name, "exec")
    trans = Transformer(file_name, src)
    patched_tree = trans.visit(tree)
    ast.fix_missing_locations(patched_tree)

    code = compile(patched_tree, file_name, "exec")

    module = ModuleType(module_name)
    setattr(module, ACTION_NAME, action_fn)
    exec(code, module.__dict__)

    return module


def exports(code_dir, file_name):
    full_path = code_dir / file_name
    with open(full_path) as fp:
        code = fp.read()

    tree = ast.parse(code, file_name, "exec")
    for node in tree.body:
        if not isinstance(node, (ast.FunctionDef, ast.ClassDef)):
            continue

        yield node.name
