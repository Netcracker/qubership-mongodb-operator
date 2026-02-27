#!/bin/bash
set -e

# Fix ownership every time on startup (optional, helpful if using mounted volumes)
chown -R mongodb:mongodb /data/db /data/configdb

# Run mongod as user 'mongodb' (UID 1001)
exec gosu mongodb "$@"