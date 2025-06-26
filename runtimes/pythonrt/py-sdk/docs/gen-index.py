#!/usr/bin/env python

from pathlib import Path
from argparse import ArgumentParser, FileType

parser = ArgumentParser(description=__doc__)
parser.add_argument(
    "output",
    help="output requirements file",
    default="index.rst",
    type=FileType("w"),
    nargs="?",
)
args = parser.parse_args()

docs_dir = Path(__file__).parent.absolute()
ak_dir = docs_dir / "../autokitteh"

index_header = """\
.. autokitteh documentation master file, created by sphinx-quickstart & gen-index.py
   sphinx-quickstart on Wed Jul  3 11:27:22 2024.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

AutoKitteh Python SDK Documentation
===================================

.. toctree::
   :maxdepth: 2
   :caption: Contents:


Module contents
---------------

.. automodule:: autokitteh
   :members:
   :undoc-members:
   :show-inheritance:



Submodules
----------
"""

mod_header = "autokitteh.{mod} module"
mod_template = """\
{header}
{line}

.. automodule:: autokitteh.{mod}
   :members:
   :undoc-members:
   :show-inheritance:
"""

print(index_header, file=args.output)

ignored = {
    "__init__.py",
}
for mod in ak_dir.glob("*.py"):
    if mod.name in ignored:
        continue

    header = mod_header.format(mod=mod.stem)
    line = "-" * len(header)
    text = mod_template.format(mod=mod.stem, header=header, line=line)
    print(text, file=args.output)


footer = """\
Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`
"""
print(footer, file=args.output)
