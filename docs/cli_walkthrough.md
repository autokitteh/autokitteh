# CLI Walkthrough

## Introduction

This is the **second** autokitteh tutorial. It demonstrates the deployment of
a project, using the `ak` Command Line Interface, on the local server which
you've prepared in the [first tutorial](./server_setup.md).

[The third tutorial](./project_deep_dive.md) will be a deep dive into the
nitty-gritty details of projects, using low-level commands in the `ak` CLI
instead of the shortcuts featured in the this tutorial.

## Clone the Samples Repository

Explore the git repository of autokitteh sample projects:
<https://github.com/autokitteh/samples>.

Each of the project directories contains:

- A `README.md` documentation file
- An `autokitteh.yaml` manifest file
- One or more program files (code and resources)

Some sample projects focus on the capabilities and APIs of specific autokitteh
integrations.

Other sample projects are more cross-functional and goal-oriented,
demonstrating the operation and value of the platform as a whole.

Choose a sample that you want to work with.

## Project Manifest Files

The `autokitteh.yaml` file is a declarative manifest that describes the
configuration of a project:

- Project name
- Program file(s)
- autokitteh connection(s)
- Execution environments (e.g. test/prod, geographical regions, availability
  zones)
- Triggers (asyncrhnous events from connections, mapped to entry-point
  functions)

This file is a shortcut, you may also configure each of these details
separately - more on that in other tutorials.

FYI, run this command to see autokitteh's YAML file schema:

```console
ak manifest schema
```

## Define a Project

> [!IMPORTANT]
> The `paths` that are specified in the `autokitteh.yaml` file are
> relative to the location of this file.

> [!IMPORTANT]
> Don't forget to replace placeholder strings in the `autokitteh.yaml` file
> before you begin: autokitteh connection tokens, environment values, etc.

Run this command to apply all the project settings:

```console
ak manifest apply <YAML file path>
```

> [!TIP]
> You may append the global flag `-J` to print the command's output as
> formatted and indented JSON (e.g. the project's name and ID, which you may
> use in the next step).

> [!NOTE]
> This is not an atomic action: some settings may fail, or be skipped.

## Build the Program

First, instruct the server to build the program:

```console
ak project build <project name or ID>
```

The output of this command is the server's auto-generated build ID (a unique
ID with the prefix `b:`). This ID will be needed when creating a deployment.

### Optional Verification

> [!CAUTION]
> Currently, the build command does not emit warnings or errors if it fails to
> read any specified file paths. This extra step mitigates this issue.

Download a copy of the build archive file from the server:

```console
ak build download <build ID> [--output=<file path>]
```

> [!TIP]
> The default output filename is `build.akb`. You may also use `-` to redirect
> to the standard output (stdout).

Examine the archive file's contents, to ensure that all your specified paths
are indeed included:

```console
tar -t -f build.akb
```

For example, if your `autokitteh.yaml` file includes these Starlark files:

```yaml
paths:
  - file_1.star
  - file_2.star
  - file_3.star
```

Then you should see this in the archive's file list:

```console
...
runtimes/starlark/resources/file_1.star
runtimes/starlark/resources/file_2.star
runtimes/starlark/resources/file_3.star
...
```

> [!NOTE]
> You don't have to keep a local copy of this build archive file for any
> reason other than this verification step. Feel free to delete it afterwards.

## Deploy the Program

First, create a new deployment:

```console
ak deployment create <--env=...> <--build_id=...>
```

The environment may be specified by its full name (`project_name.env_name`) or
ID (a unique ID with the prefix `e:`).

The build ID is a unique ID with the prefix `b:`.

For example: if the project name is `foo_bar` and its environment name
is `prod`, and the build ID from the output of the `ak project build` command
above is `b:1234567890abcdef`, then you should run this command:

```console
ak deployment create --env=foo_bar.prod --build_id=b:1234567890abcdef
```

The output of this command is the server's auto-generated deployment ID
(a unique ID with the prefix `d:`). This is needed in the next step.

Then, activate this deployment:

```console
ak deployment activate <deployment ID>
```

For verification, check the deployment's state:

```console
ak deployment get <deployment ID> -J
```

The output of this command includes:

- Deployment, build, and environment IDs
- Creation and last-update timestamps
- Current state (which should be `ACTIVE`)

## Track Program Sessions

List all sessions, in short-form:

```console
ak session list [optional filter flags] | jq 'del(.inputs)'
```

Get a single session's configuration details (entry-point, inputs, etc.):

```console
ak session get [session ID] -J
```

Get a session's runtime data (states, calls, prints, errors):

```console
ak session history [session ID] -J
```

> [!NOTE]
> In both `get` and `history` commands, the default session ID is the one with
> the latest creation timestamp. This default is useful only in single-user
> non-production servers.
