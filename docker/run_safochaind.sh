#!/bin/sh

if test -n "$1"; then
    # need -R not -r to copy hidden files
    cp -R "$1/.saf" /root
fi

mkdir -p /root/log
safrochaind start --rpc.laddr tcp://0.0.0.0:26657 --minimum-gas-prices 0.0001usaf --trace
