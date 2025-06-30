#!/usr/bin/env bash
set -e

# === Constants ===
KAFKA_HOST=127.0.0.1
KAFKA_PORT=9092
KAFKA_BROKER="${KAFKA_HOST}:${KAFKA_PORT}"

# List of topics to create
LIFECYCLE_TOPICS="backbeat-lifecycle-bucket-tasks backbeat-lifecycle-object-tasks"
NOTIFICATION_TOPICS="backbeat-bucket-notification"

function create_topics() {
    local topics="$1"
    for topic in $topics; do
        echo "[setup] Creating topic: $topic"
        kafka-topics.sh \
            --create \
            --if-not-exists \
            --bootstrap-server "$KAFKA_BROKER" \
            --topic "$topic" \
            --partitions 1 \
            --replication-factor 1
        echo "[setup] Topic '$topic' created (or already exists)."
        echo
    done
}

# === Environment Echo ===
echo "[setup] Kafka broker: $KAFKA_BROKER"
echo "[setup] Topics to create: $TOPICS"
echo

# === Wait for Kafka to be ready ===
echo "[setup] Waiting for Kafka broker..."
until nc -z "$KAFKA_HOST" "$KAFKA_PORT"; do
  echo "[setup] Kafka not ready, retrying in 2s..."
  sleep 2
done
echo "[setup] Kafka is ready!"
echo

# === Create topics ===
echo "[setup] Creating lifecycle topics..."
create_topics "$LIFECYCLE_TOPICS"

echo "[setup] Creating notification topics..."
create_topics "$NOTIFICATION_TOPICS"

echo "[setup] Creating Zookeeper paths..."
zookeeper-shell.sh localhost:2181/backbeat <<EOF
create /
create /bucket-notification
create /bucket-notification/raft-id-dispatcher
create /bucket-notification/raft-id-dispatcher/owners
create /bucket-notification/raft-id-dispatcher/leaders
create /bucket-notification/raft-id-dispatcher/provisions
create /bucket-notification/raft-id-dispatcher/provisions/1
quit
EOF

# === Done ===
echo "[setup] Setup completed successfully 🎉"
