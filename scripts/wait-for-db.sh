#!/bin/bash
# wait-for-db.sh - Wait for PostgreSQL to be ready

set -e

host="$1"
port="$2"
shift 2

until pg_isready -h "$host" -p "$port" -U "$DB_USER"; do
  echo "Waiting for PostgreSQL at $host:$port..."
  sleep 2
done

echo "PostgreSQL is ready!"
exec "$@"
