# autokitteh

autokitteh is an **open-source platform** that lets you **develop and manage**
**automated workflows** with **simple tools and familiar languages**, no
matter how complex your needs are.

autokitteh **takes care of the toil** and provides **advanced engineering**
**out-of-the-box**:

- Secure, seamless, bidirectional API integration
- User-friendly monitoring and debugging
- Standalone and distributed system reliability
- Automated recovery without state loss
- Built-in durability for long-running workflows
- Readiness for world-class scalability needs
- Versatile deployment strategies

autokitteh promotes a developer-first approach, **catering to both**
**inexperienced beginners and busy experts**. Its versatility accommodates a
wide array of use-cases, including:

- CI/CD and DevOps processes
- Infrastructure orchestration
- Ops and cybersecurity runbooks
- Cross-system syncs and integrations
- Sales, marketing, and other corp automations

## Build From Source

```
$ git clone https://github.com/autokitteh/autokitteh.git
$ cd autokitteh
$ make ak
$ ./bin/ak version
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
- [autokitteh.com](https://autokitteh.com)
- [Discord](https://discord.gg/UhnJuBarZQ)
