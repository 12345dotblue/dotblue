#!/usr/bin/env sh
set -eu

psql -v ON_ERROR_STOP=1 \
  --username "$POSTGRES_USER" \
  --dbname postgres \
  -v casdoor_db_user="$CASDOOR_DB_USER" \
  -v casdoor_db_password="$CASDOOR_DB_PASSWORD" \
  -v casdoor_db_name="$CASDOOR_DB_NAME" \
  -v dotblue_db_user="$DOTBLUE_DB_USER" \
  -v dotblue_db_password="$DOTBLUE_DB_PASSWORD" \
  -v dotblue_db_name="$DOTBLUE_DB_NAME" <<'EOSQL'
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'casdoor_db_user', :'casdoor_db_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'casdoor_db_user')\gexec
SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', :'casdoor_db_user', :'casdoor_db_password')\gexec

SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'dotblue_db_user', :'dotblue_db_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'dotblue_db_user')\gexec
SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', :'dotblue_db_user', :'dotblue_db_password')\gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'casdoor_db_name', :'casdoor_db_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'casdoor_db_name')\gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'dotblue_db_name', :'dotblue_db_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'dotblue_db_name')\gexec
EOSQL
