# Python Runtime

Implementation of Python runtime


## Patching User Code

The Python code loads the user code and patches every function call 

## Flow of Things

A run calls start a Python server with:
- Tar file containing user code
- Entry point (e.g. `review.py:on_github_pull_request`)


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
The reason do this is that `ak_runner.py` should not have any external dependencies outside of the standard library.
Once we introduce an external dependency, it will conflict with the user dependencies.

The message payload is handled by Python and is opaque to AutoKitteh.
Currently it's base64 of a pickle.

### Integration Testing

If you run `ak` with a database, then run `./testdata/create-workflow.sh` once. 
Otherwise run it every time.
This will create a deployment for `testdata/simple/`

Then run `make run-workflow`


### `ak` with database

Look for the `config.yaml` in `ak config where` directory. Then add the following

```yaml
db:
  dsn: /tmp/ak.db  # Pick any other location
  type: sqlite
```
