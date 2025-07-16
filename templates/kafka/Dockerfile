FROM alpine:latest

RUN apk add --no-cache \
    openjdk17-jre \
    bash

RUN wget https://archive.apache.org/dist/kafka/3.4.0/kafka_2.13-3.4.0.tgz \
    && tar -xzf kafka_2.13-3.4.0.tgz \
    && mv kafka_2.13-3.4.0 /opt/kafka \
    && rm kafka_2.13-3.4.0.tgz

COPY --chmod=755 setup.sh /usr/local/bin/setup-kafka.sh

ENV PATH="/opt/kafka/bin:$PATH"

WORKDIR /opt/kafka
