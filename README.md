<div align="center">

![Top banner](/docs/images/banner.jpg)

AutoKitteh is a **developer** platform for workflow automation and
orchestration. It is an easy-to-use, code-based alternative to no/low-code
platforms (such as Zapier, Workato, Make.com, 8n8) with unlimited flexibility.

In addition, it is a **durable execution** platform for long-running and
reliable workflows. It is based on [Temporal](https://temporal.io/), hiding
many of its infrastructure and coding complexities.

AutoKitteh can be self-hosted, and has a cloud offering as well.

Once installed, AutoKitteh is a scalable "serverless" platform (with batteries
included) for DevOps, FinOps, MLOps, SOAR, productivity tasks, critical
backend business processes, and more.

<br/>

![GitHub License](https://img.shields.io/github/license/autokitteh/autokitteh)
[![Go Reference](https://pkg.go.dev/badge/go.autokitteh.dev/autokitteh.svg)](https://pkg.go.dev/go.autokitteh.dev/autokitteh)
[![Go Report Card](https://goreportcard.com/badge/go.autokitteh.dev/autokitteh)](https://goreportcard.com/report/go.autokitteh.dev/autokitteh)

[![YouTube Channel](https://img.shields.io/badge/autokitteh-ff0000?logo=youtube)](https://www.youtube.com/@autokitteh-mo5sb)
[![LinkedIn](https://img.shields.io/badge/autokitteh-0e76a8?logo=linkedin)](https://www.linkedin.com/company/autokitteh/posts/?feedView=all)
[![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/autokitteh/autokitteh)](https://github.com/autokitteh/autokitteh/commits/main)

</div>

## High-Level Architecture

![Architecture diagram](/docs/images/architecture.png)

**Platform:** A scalable server that provides interfaces for building projects
(workflows), deploying them, triggering the code with webhooks or schedulers,
executing the code as durable workflows, and managing these workflows.

**API:** AutoKitteh is an API-first platform. All services are available via
gPRC / HTTP.

**Built-in integrations:** Slack, GitHub, Twilio, ChatGPT, Gemini, Gmail,
Calendar, HTTP, gRPC and many more. It's easy to add new integrations.

**Supported programming languages:** Python, Starlark (a dialect of Python),
and TypeScript (coming soon).

[Discover how it works](https://docs.autokitteh.com/how_it_works)
(in detail).

## User Interfaces

- Command Line Interface

- Visual Studio Code Extension - Build and manage workflows

  ![VS Code screenshot](/docs/images/vscode.jpg)

- Web UI

  ![Web UI screenshot](/docs/images/web_ui.jpg)

## Why You Should Give AutoKitteh a Test Drive

AutoKitteh provides a full set of advanced engineering features
out-of-the-box. You can focus on writing the business logic, we take care of
the rest:

- Secure, seamless, bidirectional API integrations
- User-friendly management, monitoring, and debugging
- Standalone and distributed system reliability
- Automated recovery without state loss
- Built-in durability for long-running workflows
- Readiness for world-class scalability needs

## Getting Started

Visit our [docs](https://docs.autokitteh.com/) here:

- [Installation](https://docs.autokitteh.com/get_started/install)
- [Starting a self-hosted server](https://docs.autokitteh.com/get_started/start_server)
- [CLI quickstart guide](https://docs.autokitteh.com/get_started/client/cli/quickstart)
- [VS Code extension](https://docs.autokitteh.com/get_started/client/vscode)

The open-source AutoKitteh server is used mostly for self-hosted and on-prem
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
- atlasgo (for new DB migrations)
- nodejs >= 20 (only if updating the UI)

## Contact Us

- meow@autokitteh.com
- https://autokitteh.com
- [Discord](https://discord.gg/UhnJuBarZQ)
