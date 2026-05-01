FROM golang:alpine as builder

# Add ca-certs
RUN apk add --update --no-cache ca-certificates

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOAMD64=v2

ADD . /build
WORKDIR /build

RUN go mod download && \
    go build -a -ldflags '-extldflags "-static"' -o example-exporter example-exporter.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/example-exporter /bin/example-exporter

ENTRYPOINT [ "/bin/example-exporter" ]