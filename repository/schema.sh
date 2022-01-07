#!/bin/sh
set -e
docker exec -it irc-calc_postgres_1 pg_dump -s irc_calcs > schema.sql
