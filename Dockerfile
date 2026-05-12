# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.1

FROM golang:${GO_VERSION}-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN for bin in jobstate requestapi executor reconciler; do \
      CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
      go build -trimpath -ldflags="-s -w" -o /out/${bin} ./cmd/${bin}; \
    done

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/jobstate /usr/local/bin/jobstate
COPY --from=build /out/requestapi /usr/local/bin/requestapi
COPY --from=build /out/executor /usr/local/bin/executor
COPY --from=build /out/reconciler /usr/local/bin/reconciler

USER nonroot:nonroot
