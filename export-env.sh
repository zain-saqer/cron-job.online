#!/bin/sh

## Usage:
##   . ./export-env.sh ; $COMMAND
##   . ./export-env.sh ; echo ${MINIENTREGA_FECHALIMITE}

uname_str=$(uname)
if [ "$uname_str" = 'Linux' ]; then

  export $(grep -v '^#' .env | xargs -d '\n')

elif [ "$uname_str" = 'FreeBSD' ] || [ "$uname_str" = 'Darwin' ]; then

  export $(grep -v '^#' .env | xargs -0)

fi