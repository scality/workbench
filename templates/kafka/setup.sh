#!/usr/bin/env bash
set -ex

if [[ -z "$KAFKA_HOST" ]]; then
    KAFKA_HOST=127.0.0.1
fi

if [[ -z "$KAFKA_PORT" ]]; then
    KAFKA_PORT=9092
fi

KAFKA_BROKER="${KAFKA_HOST}:${KAFKA_PORT}"

echo "[setup] Kafka broker: $KAFKA_BROKER"

echo "[setup] Waiting for Kafka broker..."
until nc -z "$KAFKA_HOST" "$KAFKA_PORT"; do
  echo "[setup] Kafka not ready, retrying in 2s..."
  sleep 2
done
echo "[setup] Kafka is ready!"
echo


if [[ -n "$TOPICS_TO_CREATE" ]]; then
    echo "[setup] Topics to create: $topic"
    echo
    common_opts="--create --if-not-exists --bootstrap-server $KAFKA_BROKER --partitions 1 --replication-factor 1"
    if [[ -n "$JAAS_CONFIG" ]]; then
        common_opts="$common_opts --command-config $JAAS_CONFIG"
    fi
    for topic in $TOPICS_TO_CREATE; do
        echo "[setup] Creating topic: $topic"
        kafka-topics.sh \
            $common_opts \
            --topic "$topic" \
        echo "[setup] Topic '$topic' created (or already exists)."
        echo
    done
fi

if [[ "$CREATE_NOTIFICATION_PATHS" == "true" ]]; then
    if [[ -z "$ZOOKEEPER_ENDPOINT" ]]; then
        echo "[setup] Zookeeper endpoint not set"
        exit 1
    fi

    echo "[setup] Creating Zookeeper paths..."
    zookeeper-shell.sh localhost:2181/backbeat <<EOF
create /
create /bucket-notification
create /bucket-notification/raft-id-dispatcher
create /bucket-notification/raft-id-dispatcher/owners
create /bucket-notification/raft-id-dispatcher/leaders
create /bucket-notification/raft-id-dispatcher/provisions
create /bucket-notification/raft-id-dispatcher/provisions/1
create /bucket-notification/raft-id-dispatcher/provisions/2
create /bucket-notification/raft-id-dispatcher/provisions/3
quit
EOF
    echo "[setup] Zookeeper paths created."
    echo
fi
