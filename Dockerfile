# ------------------------------------------------------------------------------
# Builder Image
# ------------------------------------------------------------------------------
FROM golang AS build

WORKDIR /build

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download
COPY ./Makefile ./Makefile

COPY ./connectivity ./connectivity
COPY ./graphcall ./graphcall
COPY ./cmd/common ./cmd/common
COPY ./cmd/manager ./cmd/manager
COPY ./cmd/manager-migration ./cmd/manager-migration
COPY ./manager ./manager

ENV CGO_ENABLED=0
ENV GOARCH=amd64
ENV GOOS=linux

RUN make build build-migration

# ------------------------------------------------------------------------------
# Target Image
# ------------------------------------------------------------------------------
FROM alpine  AS release

RUN addgroup --gid 1234 figment
RUN adduser --system --uid 1234 figment
USER 1234
WORKDIR /app

COPY --from=build  /build/cmd/manager-migration/migrations /app/migrations
COPY --from=build  /build/manager_migration_bin /app/migration

COPY --from=build  /build/manager_bin /app/manager
CMD ["/app/manager"]
