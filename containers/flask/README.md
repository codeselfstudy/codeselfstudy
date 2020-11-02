I used [this guide](https://www.digitalocean.com/community/tutorials/how-to-build-and-deploy-a-flask-application-using-docker-on-ubuntu-18-04) to help with setting up Flask with Docker, but I didn't use sudo on anything. (**Tip:** you can set up Docker so that it doesn't need `sudo`.)

Alternatively, you can reload the app with this:

```text
$ docker container stop docker.mercury && docker container start docker.mercury
```

## Old Notes

For development on a machine that doesn't have Postgres installed:

```text
$ dc run --rm \
    -p 5432:5432 \
    -v $(pwd)/pg:/var/lib/postgresql/data \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    --name mercury-postgres postgres:12-alpine
```

**WARNING:** the previous version didn't persist the volume. There must be an error with the syntax or path. I've tried updating it to `/data/db` but haven't verified that it works yet. I think it will work though.

```text
$ dc run --rm \
    --name mercury-mongo \
    -v $(pwd)/mongo:/data/db \
    -p 27017:27017 \
    mongo
```
