FROM golang:1.21 as build

COPY . /build

WORKDIR /build

ENV CGO_ENABLED=0

RUN go mod download
RUN make bin

FROM scratch

WORKDIR /bin

COPY --from=build /build/bin/ak /bin/ak

CMD ["/bin/ak", "up"]
