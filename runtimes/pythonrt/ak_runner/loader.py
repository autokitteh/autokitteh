import ast
import builtins
from pathlib import Path
from types import ModuleType

from . import log


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
        print(str(node))
        print(str(name))

        if not name or name in BUILTIN:
            return node

        log.info("%s:%d: patching %s with action", self.file_name, node.lineno, name)
        call = ast.Call(
            func=ast.Name(id=ACTION_NAME, ctx=ast.Load()),
            args=[node.func] + node.args,
            keywords=node.keywords,
        )
        return call


def load_code(root_path, action_fn, module_name):
    """Load user code into a module, instrumenting function calls."""
    log.info("importing %r", module_name)
    file_name = Path(root_path) / (module_name + ".py")
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
