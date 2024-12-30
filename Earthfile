VERSION 0.8

FROM tochemey/docker-go:1.23.4-5.1.0

pbs:
    BUILD +protogen

test:
  BUILD +lint
  BUILD +local-test

code:
    WORKDIR /app

    # download deps
    COPY go.mod go.sum ./
    RUN go mod download -x

    # copy in code
    COPY --dir . ./

vendor:
    FROM +code

    RUN go mod tidy && go mod vendor
    SAVE ARTIFACT /app /files


lint:
    FROM +vendor

    COPY .golangci.yml ./

    # Runs golangci-lint with settings:
    RUN golangci-lint run --timeout 10m


local-test:
    FROM +vendor

    WITH DOCKER --pull postgres:11
        RUN go test -mod=vendor ./... -race -v -coverprofile=coverage.out -covermode=atomic -coverpkg=./...
    END

    SAVE ARTIFACT coverage.out AS LOCAL coverage.out


protogen:
    # copy the proto files to generate
    COPY --dir protos/ ./
    COPY buf.work.yaml buf.gen.yaml ./

    # generate the pbs
    RUN buf generate \
            --template buf.gen.yaml \
            --path protos/test

    # save artifact to
    SAVE ARTIFACT gen gen AS LOCAL test/data

