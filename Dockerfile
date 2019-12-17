FROM alpine:3.10

RUN apk add --no-cache ca-certificates

ADD ./loki-operator /loki-operator

ENTRYPOINT ["/loki-operator"]
