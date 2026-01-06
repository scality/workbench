#!/usr/bin/env sh
set -e
set -x

echo "[clickhouse-setup] Starting schema initialization..."

# Wait for both shards to be ready
echo "[clickhouse-setup] Waiting for shard 1..."
until clickhouse-client --host 127.0.0.1 --port 9002 --query "SELECT 1" > /dev/null 2>&1; do
  echo "[clickhouse-setup] Shard 1 not ready, waiting 2s..."
  sleep 2
done
echo "[clickhouse-setup] Shard 1 is ready!"

echo "[clickhouse-setup] Waiting for shard 2..."
until clickhouse-client --host 127.0.0.1 --port 9003 --query "SELECT 1" > /dev/null 2>&1; do
  echo "[clickhouse-setup] Shard 2 not ready, waiting 2s..."
  sleep 2
done
echo "[clickhouse-setup] Shard 2 is ready!"

# Execute SQL files on both shards
for sql_file in /opt/init.d/*.sql; do
  filename=$(basename "$sql_file")
  echo "[clickhouse-setup] Executing $filename on shard 1..."
  clickhouse-client --host 127.0.0.1 --port 9002 --multiquery < "$sql_file"

  echo "[clickhouse-setup] Executing $filename on shard 2..."
  clickhouse-client --host 127.0.0.1 --port 9003 --multiquery < "$sql_file"
done

echo "[clickhouse-setup] Schema initialization completed successfully!"
