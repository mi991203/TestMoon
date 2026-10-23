#!/bin/sh
set -eu
while IFS= read -r database; do
  case "$database" in ''|*[!a-z_]*) echo 'Invalid database name' >&2; exit 2;; esac
  mysql --host=mysql --user=root <<SQL
CREATE DATABASE IF NOT EXISTS $database CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL ON $database.* TO 'testmoon'@'%';
SQL
done < /init/databases.txt
