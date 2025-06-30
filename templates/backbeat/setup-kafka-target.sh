#!/usr/bin/env bash
set -e

# === Constants ===
KAFKA_HOST=127.0.0.1
KAFKA_PORT=9093
KAFKA_BROKER="${KAFKA_HOST}:${KAFKA_PORT}"

topic="destination-topic"

# === Environment Echo ===
echo "[setup] Kafka broker: $KAFKA_BROKER"
echo "[setup] Topics to create: $topic"
echo

# === Wait for Kafka to be ready ===
echo "[setup] Waiting for Kafka broker..."
until nc -z "$KAFKA_HOST" "$KAFKA_PORT"; do
  echo "[setup] Kafka not ready, retrying in 2s..."
  sleep 2
done
echo "[setup] Kafka is ready!"
echo

echo "[setup] Creating topic: $topic"
kafka-topics.sh \
    --create \
    --if-not-exists \
    --bootstrap-server "127.0.0.1:9093" \
    --topic "$topic" \
    --partitions 1 \
    --replication-factor 1
echo "[setup] Topic '$topic' created (or already exists)."
echo
