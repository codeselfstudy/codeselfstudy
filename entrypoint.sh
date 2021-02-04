#!/bin/ash

# Instead of putting the symlinks in the repo, create them in the container?
# if [ ! -L ./containers/gatsby/static/files ] && [ ! -e ./containers/gatsby/static/files ]; then
    # create symlink inside the container here?
# fi

# See: https://medium.com/ultralight-io/getting-gatsby-to-run-on-docker-compose-6bf3b0d97efb
ln -s /save/node_modules/* ./node_modules/.
gatsby develop -H 0.0.0.0
