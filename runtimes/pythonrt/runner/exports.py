from argparse import ArgumentParser
import json
import sys

import loader
from main import dir_type


async def main():
    parser = ArgumentParser(description="print exports in JSON format")
    parser.add_argument("code_dir", help="code directory", type=dir_type)
    args = parser.parse_args()

    exports = []
    for path in args.code_dir.glob("*.py"):
        if path.name[0] == ".":
            continue

        async for e in loader.exports(args.code_dir, path.name):
            exports.append(e)

    json.dump(exports, sys.stdout)


if __name__ == "__main__":
    import asyncio

    asyncio.run(main())
