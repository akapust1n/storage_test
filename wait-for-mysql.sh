#!/bin/bash
# wait-for-mysql.sh

set -e

host="${MYSQL_HOST:-db:3306}"
user="${MYSQL_USER:-filestore}"
password="${MYSQL_PASSWORD:-filestore_password}"
database="${MYSQL_DATABASE:-filestore}"

until mysqladmin ping -h"${host%:*}" -u"$user" -p"$password" --silent; do
  echo "Waiting for MySQL to be ready..."
  sleep 1
done

# Дополнительная проверка готовности базы данных
echo "Checking MySQL connection..."
for i in {1..30}; do
  if mysql -h"${host%:*}" -u"$user" -p"$password" -e "SELECT 1" >/dev/null 2>&1; then
    echo "MySQL is fully operational!"
    break
  fi
  echo "Waiting for MySQL to be fully operational... ($i/30)"
  sleep 1
done

echo "MySQL is ready! Starting application..."
exec "$@" 