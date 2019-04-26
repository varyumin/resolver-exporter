FROM golang:1.12 as builder
RUN useradd -u 1111 app
COPY src/main.go /tmp/resolver/
WORKDIR /tmp/resolver/
RUN go mod init github.com/varyumin/resolver && \
    go build -o /tmp/resolver/resolver

FROM ubuntu
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /tmp/resolver/resolver /go/bin/resolver_exporter
RUN chown -R app:app /go
USER app
CMD ["/go/bin/resolver_exporter"]