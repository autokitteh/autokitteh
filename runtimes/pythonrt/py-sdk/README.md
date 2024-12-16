# AutoKitteh Python SDK

This is the Python SDK for [AutoKitteh](https://autokitteh.com).

It provides the following utilities for AutoKitteh workflows:

- Client initialization functions for AutoKitteh connections
- Un/subscription and consumption functions for AutoKitteh events
- `activity` decorator to work around issues with [pickle](https://docs.python.org/3/library/pickle.html)

For more information, see [our documentation](https://docs.autokitteh.com/develop/python).

## Installing

Installing is required only for local testing.

When AutoKitteh runs your workflows it'll install this library automatically.

```
python -m pip install autokitteh
```

## Building the Documentation

Online documentation is at https://autokitteh.readthedocs.io
The documentation is at the `docs` directory.

To make sure readthedocs have all the required dependencies, run `gen-reqs.py` in the `docs` directory and commit `requirements.txt`
