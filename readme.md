# 🎒GoPack

[![build](https://img.shields.io/github/actions/workflow/status/Tochemey/gopack/build.yml?branch=main)](https://github.com/Tochemey/gopack/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/Tochemey/gopack/branch/main/graph/badge.svg?token=LJO3LHe1Ox)](https://codecov.io/gh/Tochemey/gopack)

GoPack is kind of your Swiss Army Knife for golang services.
The project adheres to [Semantic Versioning](https://semver.org)
and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## Features

- [gRPC](./grpc) - contains client and server
    - Traces and Metrics are automatically handled depending upon the configuration.
    - ratelimiter interceptors (unary/stream) for both client and server
    - trace interceptors (unary/stream) for both client and server
    - metrics interceptors (unary/stream) for both client and server
    - recovery interceptors (unary/stream) for both client and server
    - request id interceptors (unary/stream) for both client and server
    - customizable options for both gRPC client and server
    - testkit to start a gRPC test server
- [Postgres](./postgres) - contains postgres database interface to execute SQL statement with postgres with traces and metrics out of the box.
    - testkit to smoothly implement unit/integration tests with postgres
- [OpenTelemetry](./otel) - contains trace and metrics provider to simplify the creation of trace and metrics providers.
    - testkit to create an opentelemetry test collector
- [Scheduler](./scheduler) - contains a crontab library to implement job schedulers.
- [Zap logger](./log/zapl) wrapper with _request_id_, _trace_id_ and _span_id_ injected to log when present in the context
- [Request ID](./requestid) - contains request ID injection with appropriate interceptors and handlers.
- [Validation](./validation) - contains a simple validation library.
- [Errors Chain](./errorschain) - contains an simple errors chain library.
- [Stream](./stream) - contains a simple in-memory pubsub
- [Ticker](./ticker) - contains an enhanced golang ticker
- [Timer Pool](./timerpool) - contains a timer pool for memory efficiency
- [Collection](./collection) - contains some thread-safe collections like map, slice and queue
- [Future](./future) - contains a future library to handle async calls
- [GCP PubSub](./gcp/pubsub) - contains wrappers around the Google PubSub api to:
  - [Tooling](./gcp/pubsub/tooling.go) - create, list topics
  - [Publisher](./gcp/pubsub/publisher.go) - publish messages to [GCP PubSub](https://cloud.google.com/pubsub/docs)
  - [Subscriber](./gcp/pubsub/subscriber.go) - consume messages from [GCP PubSub](https://cloud.google.com/pubsub/docs)
  - [Emulator](./gcp/pubsub/emulator.go) - contains a [GCP PubSub](https://cloud.google.com/pubsub/docs) Emulator.

### Note

Traces and Metrics are accessible via the integration
with [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go).

## Contribution

Contributions are welcome!
The project adheres to [Semantic Versioning](https://semver.org)
and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).
This repo uses [Earthly](https://earthly.dev/get-earthly).

To contribute please:

- Fork the repository
- Create a feature branch
- Submit a [pull request](https://help.github.com/articles/using-pull-requests)

### Test & Linter

Prior to submitting a [pull request](https://help.github.com/articles/using-pull-requests), please run:

```bash
earthly +test
