#!/bin/sh

export DIPS_HOST="${DIPS_HOST:-rabbitmq:rabbitmq@172.17.0.1}"

export MONGODB_HOST="${MONGODB_HOST:-mongodb://172.17.0.1:27017}"
export MONGODB_AUTH_MECHANISM="${MONGODB_AUTH_MECHANISM:-SCRAM-SHA-256}"
export MONGODB_AUTH_SOURCE="${MONGODB_AUTH_SOURCE:-dips}"
export MONGODB_USERNAME="${MONGODB_USERNAME:-dips}"
export MONGODB_PASSWORD="${MONGODB_PASSWORD:-dips}"
export MONGODB_DATABASE="${MONGODB_DATABASE:-dips}"

export INFLUXDB_HOST="${INFLUXDB_HOST:-http://172.17.0.1:8086}"
export INFLUXDB_DATABASE="${INFLUXDB_DATABASE:-dips}"
export INFLUXDB_USERNAME="${INFLUXDB_USERNAME:-dips}"
export INFLUXDB_PASSWORD="${INFLUXDB_PASSWORD:-dips}"

cat << EOT > ./config.yml
dips:
  host: "$DIPS_HOST"

mongodb:
  hosts:
    - "$MONGODB_HOST"
  auth_mechanism: "$MONGODB_AUTH_MECHANISM"
  auth_source: "$MONGODB_AUTH_SOURCE"
  username: "$MONGODB_USERNAME"
  password: "$MONGODB_PASSWORD"
  database: "$MONGODB_DATABASE"

influxdb:
  host: "$INFLUXDB_HOST"
  database: "$INFLUXDB_DATABASE"
  username: "$INFLUXDB_USERNAME"
  password: "$INFLUXDB_PASSWORD"
EOT

exec "$@"
