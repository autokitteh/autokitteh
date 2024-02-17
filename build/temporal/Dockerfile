FROM ubuntu:20.04 as base

RUN apt update -y && apt install curl -y

RUN curl -sSf https://temporal.download/cli.sh | sh


FROM gcr.io/distroless/base-debian11

WORKDIR /usr/src/app

COPY --from=base /root/.temporalio/bin/temporal temporal

ENTRYPOINT ["./temporal", "server", "start-dev", "-n", "default", "--ip" , "0.0.0.0"]
