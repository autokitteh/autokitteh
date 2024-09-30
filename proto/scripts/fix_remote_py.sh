#!/bin/bash

# Make directories in pb importable
find ../runtimes/pythonrt/runner/pb -type d -exec touch {}/__init__.py \;

