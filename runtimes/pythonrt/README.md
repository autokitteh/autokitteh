# Python Runtime

Implementation of Python runtime. 
See [Python runtime](https://linear.app/autokitteh/project/python-runtime-be87fe4c4d7d) for list of issues.

Currently, we don't support 3rd party packages (e.g. `pip install`) for the user code.
See [ENG-538](https://linear.app/autokitteh/issue/ENG-538/support-python-dependencies) for more details.
For a realistic POC/demo, we'll pre-install the packages the user code requires (e.g. `slack-sdk`).

## Python Handler Function

Python's handler functions should receive a single `dict` value which is the triggering event.
This event has the following keys:

- `event_type` (str): The event type
- `event_id` (str): The event ID
- `original_event_id` (str): The original event ID
- `integration_id` (str): The ID of the integration that fired the event
- `data` (dict): The event payload (integration specific)

The return value from Python entry points is ignored and is not passed back to autokitteh.

## Patching User Code

The Python code (`ak_runner.py`) loads the user code and patches every function call.
It does so by hooking into the regular import hook.
When the user module is loaded, we transform the AST to change function calls from:

```python
urlopen(url)
```

to:

```python 
_ak_call(urlopen, url)
```

`ak_call` will call the Go Python runtime which will start an activity.

### Detecting External Function Calls

When the user module is loaded, we set `MODULE_NAME` the name of the loaded module.
When `ak_call` is invoked with a function it compares the function module name with `MODULE_NAME`,
if they are the same, it's an internal function and `ak_call` will return the invocation of the function.
Otherwise, `ak_call` will call the Go process that will invoke an activity.

See [ENG-495](https://linear.app/autokitteh/issue/ENG-495/better-detection-of-external-functions).

## Go ↔ Python Communication Flow

A run calls start a Python server with:
- Tar file containing user code
- Entry point (e.g. `review.py:on_github_pull_request`)

It will also inject `vars` definition from the manifest to the Python process environment.

The Python server returns a list of exported symbols from the user code.

### Communication sequence

A call with function and payload:

```
Go                              Python

------ Call (function, payload) ------->

<--- Activity request (payload) ----
----> Activity call (payload) --->
<---- Activity result (value) -----
----- Activity result  ----->

<--- Activity request (payload) ----
----> Activity call (payload) --->
<---- Activity result (value) -----
----- Activity result  ----->

...

<------ Call result (value) ----

```

### Communication Protocol

We're using JSON over Unix domain socket, one JSON object per line.
The reason do this is that `ak_runner.py` should not have any external dependencies outside the standard library.
Once we introduce an external dependency, it will conflict with the user dependencies.

The message payload is handled by Python and is opaque to autokitteh.
Currently, it's base64 of a pickle.


### Integration Testing

If you run `ak` with a database, then run `make create-workflow` once. 
Otherwise run it every time.
This will create a deployment for `testdata/simple/`

Then run `make run-workflow`.

### `ak` with database

Look for the `config.yaml` in `ak config where` directory. Then add the following

```yaml
db:
  dsn: /tmp/ak.db  # Pick any other location
  type: sqlite
```
