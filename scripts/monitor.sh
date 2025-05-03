#!/bin/sh

trap _exit INT

_exit() {
  echo "*** interrupt caught, exiting... ***"
  exit
}

while true; do
  # Monitor command is blocking, we run it in a loop here as the PTY is
  # destroyed during the flasing process, so we just keep on trying to
  # reconnect until it is back up again.
  sudo tinygo monitor
  # If monitor command exited without errors, we must have control-c'd it. In
  # any case, exit.
  [[ $? == 0 ]] && _exit
  # Protect against any possible type of spinlock.
  sleep 1
done
