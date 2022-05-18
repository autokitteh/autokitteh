FROM golang:1.18

COPY install.sh /
RUN chmod +x /install.sh && /install.sh && rm -fR /install.sh

CMD ["/usr/bin/protoc"]
