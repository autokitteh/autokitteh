[![Go Reference](https://pkg.go.dev/badge/go.autokitteh.dev/autokitteh.svg)](https://pkg.go.dev/go.autokitteh.dev/autokitteh)
[![Go Report Card](https://goreportcard.com/badge/go.autokitteh.dev/autokitteh)](https://goreportcard.com/report/go.autokitteh.dev/autokitteh)

# autokitteh

autokitteh is an open-source platform for developing and managing automated,
reliable, durable, long-running workflows with simple tools and familiar
languages.

It is an alternative to both no-code/low-code platforms (such as Zapier and
make.com) and complex durable execution systems (such as Temporal.io), with
abstractions that don't sacrifice engineering power and flexibility, no matter
how complex your needs are.

autokitteh promotes a developer-first approach, catering to both inexperienced
beginners and busy experts, with a wide variety of skill sets and use cases:

- CI/CD pipelines and DevOps processes
- Infrastructure and backend systems orchestration
- IT, ops, and cybersecurity SOAR runbooks
- Cross-system syncs and integrations
- Sales, marketing, and back-office automations

autokitteh hides away the toil and provides advanced engineering features
out-of-the-box:

- Secure, seamless, bidirectional API integration
- User-friendly management, monitoring, and debugging
- Standalone and distributed system reliability
- Automated recovery without state loss
- Built-in durability for long-running workflows
- Readiness for world-class scalability needs
- Versatile deployment strategies

Here's a [detailed look at how autokitteh works](https://docs.autokitteh.com/how_it_works).

## User Instructions

[Getting started](https://docs.autokitteh.com/get_started):

- [Installation](https://docs.autokitteh.com/get_started/install)
- [Starting a local server](https://docs.autokitteh.com/get_started/start_server)
- [Running workflows](https://docs.autokitteh.com/get_started/run_workflows)
  (with a CLI)
- [VS Code extension](https://docs.autokitteh.com/get_started/vscode_extension)

This open-source project can be used mostly for self-hosted and on-prem
installations. Our managed cloud iPaaS offering is currently in beta - for
details, contact us at meow@autokitteh.com.

## Build From Source

The following requires [Go version 1.22](https://go.dev/dl/) or greater.

```shell
$ git clone https://github.com/autokitteh/autokitteh.git
$ cd autokitteh
$ make ak
$ cp ./bin/ak /usr/local/bin
$ ak version
```

## Build Requirements (_Full_ Builds Only)

- buf
- docker
- go >= 1.22
- golangci-lint (auto-downloaded during builds if missing)
- shellcheck (auto-pulled via docker during builds if missing)

## Recommended Dev Tools

- gofumpt
- gotestsum (used by Makefile intead of "go test" if detected)
- jq (for advanced output formatting)

## Contact Us

- meow@autokitteh.com
- https://autokitteh.com
- [Discord](https://discord.gg/UhnJuBarZQ)
