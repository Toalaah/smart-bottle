#!/bin/sh

prefix="${1}"
priv="${prefix}private.pem"
pub="${prefix}public.pem"

[ -f "$priv" ] && { echo "generate: private key '${priv}' already exists, skipping regeneration"; exit 0; }

openssl genpkey -algorithm x25519 -out "$priv"
openssl pkey -in "$priv" -pubout -out "$pub"
