#!/bin/sh
set -e

# Fix permissions on data and log directories
chown -R clickhouse:clickhouse /var/lib/clickhouse
chown -R clickhouse:clickhouse /var/log/clickhouse-server

# Switch to clickhouse user and start server
exec su clickhouse -s /bin/sh -c 'exec /usr/bin/clickhouse-server --config-file=/etc/clickhouse-server/config.xml'
