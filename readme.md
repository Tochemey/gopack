# 🎒GoPack

[![build](https://img.shields.io/github/actions/workflow/status/Tochemey/gopack/build.yml?branch=main)](https://github.com/Tochemey/gopack/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/Tochemey/gopack/branch/main/graph/badge.svg?token=LJO3LHe1Ox)](https://codecov.io/gh/Tochemey/gopack)

GoPack is kind of your Swiss Army Knife for golang microservices.
The project adheres to [Semantic Versioning](https://semver.org) and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## Features
- gRPC client and server builder to quickly build and start a gRPC service. Traces and Metrics are automatically handled depending upon the configuration.
    - gRPC server to build gRPC services
    - gRPC client to build gRPC clients
    - ratelimiter interceptors (unary/stream) for both client and server
    - trace interceptors (unary/stream) for both client and server
    - metrics interceptors (unary/stream) for both client and server
    - recovery interceptors (unary/stream) for both client and server
    - request id interceptors (unary/stream) for both client and server
    - customizable options for both client and server
    - testkit to start a gRPC test server
- Postgres database interface to execute SQL statement with postgres with traces and metrics out of the box.
    - testkit to smoothly implement unit/integration tests with postgres
- OpenTelemetry trace and metrics provider to simplify the creation of trace and metrics providers
    - testkit to create a opentelemetry test collector

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
