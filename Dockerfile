FROM golang:1.27.3-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=unknown

ENV CGO_ENABLED=0

RUN go build -trimpath \
      -ldflags "-s -w \
        -X github.com/nirantarasi/k3sctl/version.Version=${VERSION} \
        -X github.com/nirantarasi/k3sctl/version.GitCommit=${GIT_COMMIT}" \
      -o /out/k3sctl .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates openssh-client tini && \
    addgroup -S k3sctl && adduser -S -G k3sctl k3sctl

COPY --from=builder /out/k3sctl /usr/local/bin/k3sctl

USER k3sctl

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD ["/usr/local/bin/k3sctl", "version"]

ENTRYPOINT ["/sbin/tini", "--", "k3sctl"]
CMD ["--help"]
