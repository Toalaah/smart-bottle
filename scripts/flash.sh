#!/bin/sh

sudo make flash "$@" &
pid=$!

while [ ! -b /dev/disk/by-label/RP2350 ]; do
  sleep 0.5
done

udisksctl mount -b /dev/disk/by-label/RP2350 2>/dev/null 1>&2
wait "$pid"
