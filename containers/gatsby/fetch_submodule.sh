#!/usr/bin/env bash

echo 'LOOK HERE'
pwd
echo '======='
pushd src/content
git submodule init
git submodule update
popd
echo 'DONE'
