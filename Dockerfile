FROM golang AS builder
ARG TARGETARCH

WORKDIR /go/src/github.com/cospotato/fnos-acme

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o bin/fnos-acme ./cmd/fnos-acme

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app/fnos-acme

COPY --from=builder /go/src/github.com/cospotato/fnos-acme/bin/fnos-acme /usr/local/bin/fnos-acme

ENTRYPOINT [ "fnos-acme", "run" ]