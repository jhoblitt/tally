FROM golang:1.19-alpine as builder

ARG BIN=tally
RUN apk --update --no-cache add \
    binutils \
    && rm -rf /root/.cache
WORKDIR /go/src/github.com/jhoblitt/tally
COPY . .
RUN go build && strip "$BIN"

FROM alpine:3
WORKDIR /root/
COPY --from=builder /go/src/github.com/jhoblitt/tally/$BIN /bin/$BIN
ENTRYPOINT ["/bin/tally"]
