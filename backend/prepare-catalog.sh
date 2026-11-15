#!/bin/sh
set -eu

if [ ! -e /data/catalog_curation.json ]; then
  cp /data/catalog_curation.example.json /data/catalog_curation.json
  chmod 600 /data/catalog_curation.json
fi

exec /catalog-export "$@"
