ARG GO_VERSION=1.25.7
FROM golang:${GO_VERSION}-alpine

ARG GOLANGCI_LINT_VERSION=v2.10.1
ARG BUF_VERSION=v1.65.0

RUN apk --no-cache add \
    bash \
    binutils-gold \
    ca-certificates \
    curl \
    docker-cli \
    gcc \
    git \
    libc-dev \
    make \
    musl-dev \
    openssh

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
 && go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest \
 && go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest \
 && GO111MODULE=on GOBIN=/usr/local/bin go install github.com/bufbuild/buf/cmd/buf@${BUF_VERSION} \
 && GO111MODULE=on GOBIN=/usr/local/bin go install github.com/vektra/mockery/v2@v2.53.2

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
        | sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}

ENV PATH="/go/bin:${PATH}"

WORKDIR /app
