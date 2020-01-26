#!/bin/ash

# See: https://medium.com/ultralight-io/getting-gatsby-to-run-on-docker-compose-6bf3b0d97efb
ln -s /save/node_modules/* ./node_modules/.
gatsby develop -H 0.0.0.0
