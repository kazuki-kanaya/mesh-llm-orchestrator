# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.1

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/jobstate ./cmd/jobstate
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/requestapi ./cmd/requestapi
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/executor ./cmd/executor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/reconciler ./cmd/reconciler

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/jobstate /usr/local/bin/jobstate
COPY --from=build /out/requestapi /usr/local/bin/requestapi
COPY --from=build /out/executor /usr/local/bin/executor
COPY --from=build /out/reconciler /usr/local/bin/reconciler

USER nonroot:nonroot
