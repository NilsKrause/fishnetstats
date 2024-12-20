#!/bin/bash -e

args=("-c /data/config.yaml")

if [ -n "$CONFIG" ]; then args+=("-c" "$CONFIG"); fi

exec /stats ${args[@]}