FROM golang:1.12.0-alpine3.9 AS build-env

RUN apk update && apk add git && apk add ca-certificates

RUN adduser -D -g '' kiwi
RUN mkdir /cryptoflux && chown kiwi /cryptoflux
WORKDIR /cryptoflux

# <- COPY go.mod and go.sum files to the workspace
COPY go.mod .
COPY go.sum .

#get dependencies
RUN go mod download

# COPY the source code
COPY . .

ENV GO111MODULE=on
# Compile the binary, we don't want to run the cgo resolver
RUN CGO_ENABLED=0 go build ./cmd/cryptoflux

# build small image from scratch
FROM scratch

VOLUME /tmp

WORKDIR /
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /cryptoflux/cryptoflux /go/bin/cryptoflux

USER kiwi

ENTRYPOINT ["/go/bin/cryptoflux"]
