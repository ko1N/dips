#!/bin/sh

export DIPS_HOST="${DIPS_HOST:-rabbitmq:rabbitmq@172.17.0.1}"

cat << EOT > ./config.yml
dips:
  host: "$DIPS_HOST"
EOT

exec "$@"
