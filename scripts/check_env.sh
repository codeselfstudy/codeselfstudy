#!/usr/bin/env bash

# This script checks to see if the user has copied `.env-example` to
# `.env`.
if [ -d "./.env" ]
then
    echo "found .env directory"
else
    echo "WARNING: did not find a .env directory. Please read the installation instructions."
    exit 1
fi
