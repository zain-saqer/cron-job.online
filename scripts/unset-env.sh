#!/bin/sh

unset $(grep -v '^#' .env | sed -E 's/(.*)=.*/\1/' | xargs)