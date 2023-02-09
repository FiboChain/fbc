#!/usr/bin/env sh

##
## Input parameters
##
ID=${ID:-0}
LOG=${LOG:-fbchaind.log}

##
## Run binary with all parameters
##
export FBCHAINHOME="/fbchaind/node${ID}/fbchaind"

if [ -d "$(dirname "${FBCHAINHOME}"/"${LOG}")" ]; then
  fbchaind --chain-id fbc-1 --home "${FBCHAINHOME}" "$@" | tee "${FBCHAINHOME}/${LOG}"
else
  fbchaind --chain-id fbc-1 --home "${FBCHAINHOME}" "$@"
fi

