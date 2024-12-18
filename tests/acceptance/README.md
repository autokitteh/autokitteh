# Acceptance Test Project

This AutoKitteh project is used as a testing tool for:

- Acceptance of new cloud environments
- Connection auth types, UIs, and flows
- Python runtime issues involving connections

It is meant to be simpler and quicker than running all the projects in
[Kittehub](https://github.com/autokitteh/kittehub/):

- Single project deployment
- Simple but comprehensive triggers
- Trivial API calls without runtime-dependent input or logic

In other words, running this project successfully is a prerequisite for testing
all the projects in [Kittehub](https://github.com/autokitteh/kittehub/).

## Expected Connection Types and Names

See the `connections` section in the [autokitteh.yaml](./autokitteh.yaml)
file.

## Connection Auth Types

For each connection, configure and re-run this project with all available auth
types.

Reminder: once a connection's auth type is set it can't be changed, so you'll
have to delete and recreate each connection in order to test all of its auth
types.

## Supported Trigger Events

See the `triggers` section in the [autokitteh.yaml](./autokitteh.yaml) file.

Note that all entry-point handler functions print the received event, but they
don't use the event's data payload.

## API Calls as Manual Runs

See the [api_calls.py](./api_calls.py) file.
