ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /usr/local/bin/telnest

FROM alpine:3.19

# add unprivileged user 
RUN addgroup -S telnest && adduser -S -G telnest telnest

COPY --from=builder /usr/local/bin/telnest /usr/local/bin/telnest

USER telnest
WORKDIR /home/telnest

EXPOSE 23/tcp
EXPOSE 2323/tcp
EXPOSE 4000/tcp

CMD ["/usr/local/bin/telnest"]
