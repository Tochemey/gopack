# 🎒GoPack

[![build](https://img.shields.io/github/actions/workflow/status/Tochemey/gopack/build.yml?branch=main)](https://github.com/Tochemey/gopack/actions/workflows/build.yml)

GoPack is kind of your Swiss Army Knife for golang microservices.
The project adheres to [Semantic Versioning](https://semver.org) and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## Features
-[x] gRPC client and server builder to quickly build and start a gRPC service. Traces and Metrics are automatically handled depending upon the configuration.
    - [x] gRPC server to build gRPC services
    - [x] gRPC client to build gRPC clients
    - [x] ratelimiter interceptors (unary/stream) for both client and server
    - [x] trace interceptors (unary/stream) for both client and server
    - [x] metrics interceptors (unary/stream) for both client and server
    - [x] recovery interceptors (unary/stream) for both client and server
    - [x] request id interceptors (unary/stream) for both client and server
    - [x] customizable options for both client and server
    - [x] testkit to start a gRPC test server
    - [ ] logging interceptors
-[x] Postgres database interface to execute SQL statement with postgres with traces and metrics out of the box.
    - [x] testkit to smoothly implement unit/integration tests with postgres
- [x] OpenTelemetry trace and metrics provider to simplify the creation of trace and metrics providers
    - [x] testkit to create a opentelemetry test collector 
- [ ] Scheduler

### Note
Traces and Metrics are accessible via the integration with [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go).

## Contribution
Contributions are welcome!
The project adheres to [Semantic Versioning](https://semver.org) and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).
This repo uses [Earthly](https://earthly.dev/get-earthly).

To contribute please:
- Fork the repository
- Create a feature branch
- Submit a [pull request](https://help.github.com/articles/using-pull-requests)

### Test & Linter
Prior to submitting a [pull request](https://help.github.com/articles/using-pull-requests), please run:
```bash
earthly +test
