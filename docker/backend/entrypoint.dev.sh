#!/bin/sh

if [ ! -f /src/.air.toml ]; then
  touch /src/.air.toml
fi

exec air -c .air.toml
