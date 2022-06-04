FROM golang:1.18 AS builder

COPY . /build
WORKDIR /build

ENV GOOS=linux

ENV GO_BUILD_OPTS="-a -ldflags '-linkmode external -extldflags "-static"'"

ARG COMMIT
ENV COMMIT=${COMMIT}

ARG VERSION
ENV VERSION=${VERSION}

ARG DATE
ENV DATE=${DATE}

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
