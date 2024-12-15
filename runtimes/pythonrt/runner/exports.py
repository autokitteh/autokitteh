# Must be first import
import filter_warnings  # noqa: F401

from argparse import ArgumentParser
import json
import sys

import loader
from main import dir_type

if __name__ == "__main__":
    parser = ArgumentParser(description="print exports in JSON format")
    parser.add_argument("code_dir", help="code directory", type=dir_type)
    args = parser.parse_args()

    exports = []
    for path in args.code_dir.glob("*.py"):
        if path.name[0] == ".":
            continue

        exports += loader.exports(args.code_dir, path.name)

    json.dump(exports, sys.stdout)
