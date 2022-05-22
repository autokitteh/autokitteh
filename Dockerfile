FROM golang:1.18 AS builder

COPY . /build
WORKDIR /build

ENV GOOS=linux

# see https://awstip.com/containerize-go-sqlite-with-docker-6d7fbecd14f0
ENV CGO_ENABLED=1
ENV GO_BUILD_OPTS="-a -ldflags '-linkmode external -extldflags "-static"'"

RUN go mod download
RUN make bin

#---

FROM alpine:latest

RUN apk update && apk --no-cache add ca-certificates bash curl

COPY --from=builder /build/bin/* /ak/bin/

WORKDIR /ak

EXPOSE 20000
EXPOSE 20001

ENTRYPOINT ["/ak/bin/akd"]
