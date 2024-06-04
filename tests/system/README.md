# Systems Tests

This directory runs end-to-end "black-box" system tests of the autokitteh CLI,
functioning both as a server and as a client.

It can also control other tools, dependencies, and in-memory fixtures (e.g.
Temporal, databases, caches, and HTTP webhooks).

Test cases are defined as [txtar](https://pkg.go.dev/golang.org/x/tools/txtar)
files in the [testdata](./testdata/) directory tree. Their structure and
scripting language is defined below.

Other than local and CI/CD testing, this may be used for benchmarking,
profiling, and load/stress testing.

## Motivation

Inspiration: <https://research.swtch.com/testing>

Guiding principles:

- Unit tests are preferable to system tests, but not always practical
- If a test case is hard to write, it won’t be be written
- If a test case is hard to read, it will be ignored, and become dead code
- Test cases should be super easy and fast to run, debug, and update
- Brittleness, nondeterminism, and lack of isolation lead to flakiness
- Tests don’t ensure correctness in the present, they codify contracts and
  assumptions, and protect them against unwitting changes in the future
- Fixing a bug? If you didn't add a test, and compare its results before and
  after the fix, you didn’t fix the bug

## Execution Cheatsheet

To run all the tests:

```
gotestsum -f testname
```

When running from the repo's root directory, you need to specify the go path:

```
gotestsum -f testsum ./systest
```

To run only a specific subset of test-case txtar files, append the flag
`-run XYZ`, where `XYZ` is a regular expression. This regular expression can
be split by slash characters (`/`) into a sequence of regular expressions,
and each part of a test's identifier (i.e. the txtar relative path under
`testdata`) must match the corresponding element in the sequence.

For example, to run only with the txtar files in/under `testdata/*subdir*/*filename*`,
including `testdata/*subdir*/*filename*.txtar`:

```
gotestsum -f testname ./tests/system -run /subdir/filename
```

To repeat a test `N` times when investigating flakiness, append the `-count`
flag:

```
gotestsum -f testname ./tests/system -run /subdir/filename -count N
```

## Txtar File Structure

The first section in every txtar file is free-form, multi-line text - this is
the test-case's script. See the [Scripting Language](#scripting-language)
section below for details.

Optional: if the script references any files, you may embed them in the txtar
file, after the script. Each file begins with the marker line `-- filename --`,
followed by zero or more file content lines.

The system test discards all leading and trailing whitespaces from sections
and filenames.

When the system test starts to run a new test-case, i.e. a new txtar file, it:

1. Parses the script section to detect syntax errors
2. Extracts all the embedded files into a new temporary directory
3. Sets up the test fixtures (e.g. AK server, HTTP server)
4. Starts running the script, line by line

## Scripting Language

Empty lines and comments (lines starting with `#`) are ignored.

"Action lines" are commands that the test executes - see details below.

Each command may have optional "customization" and/or "check" lines below it.

There's no limit on the number or repeatability of checks, e.g. you may
specify multiple `contains` or `regex` checks per command.

### Action: AK Client Command

`ak [CLI commands, arguments, and flags]`

Note: May reference filenames embedded in the same txtar file.

#### Optional Checks

`output <euqals|contains|regex> <string>`

`output <euqals|contains|regex> file <embedded txtar filename>`

`return code == <integer>`

### Action: HTTP Transaction

`http <get|post> <URL>`

- Default scheme: `http://`
- Default address: the test's AK server

**TODO:** Add a `server` action to check client requests from AK sessions, to
test integrations and sessions.

#### Optional Request Customizations

`req header <name> = <value>`

`req body <string>`

`req body file <embedded txtar filename>`

#### Optional Response Checks

`resp code == <integer>`

`resp header <name> == <value>`

`resp redirect <euqals|contains|regex> <URL>`

`resp body <euqals|contains|regex> <string>`

`resp body <euqals|contains|regex> file <embedded txtar filename>`

### Action: Wait for|unless Session

`wait <duration> <for|unless> session <session ID>`

waits up to the given duration (e.g. `5s`) for the specific session

- if wait type is `for` - waits for the session to be in one of the states `COMPLETED`, `STOPPED` or `ERROR` and report error if no session was created or it failed to reach desired state.
- if wait type is `unless` - waits for the session and reports error if found in any state.

## Syntax Summary

Actions:

```
<ak | http <get | post> | wait> <*>
```

Customizations:

```
req header <*> = <*>
req body [file] <*>
```

Checks:

```
<output | resp <body | redirect> > <euqals | contains | regex> [file] <*>
<return | resp> code == <integer>
resp header <name> == <value>
```

## Tips for Writing Tests

If you're running an AK command, and checking both its return code and output:

- When expecting success - check `return code == 0` first, and then
  `output ...`, because if the AK command failed checking the output is
  pointless
- When expecting an error - check the `output` first, and then
  `return code == <int>`, because if there are mismatches in any `output`
  check, we print the actual output, but an incorrect error code is typically
  just a minor bug or mistake
