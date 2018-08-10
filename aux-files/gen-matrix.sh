#!/bin/zsh
awk 'ORS=NR%8?" ":"\n"' | nl
