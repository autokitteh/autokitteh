FROM golang:1.18

ARG ARCH
ENV ARCH=${ARCH}

ADD scripts /scripts
RUN chmod +x /scripts/install.sh && /scripts/install.sh && rm -fR /scripts

CMD ["/usr/bin/protoc"]
