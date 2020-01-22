# Coding Puzzles

A tool to scrape and serve coding puzzles.

## Notes

You can edit the `docker-compose.yml` and `Dockerfile` as needed.

### Database

The default `docker-compose.yml` file basically replaces the need to do this:

```text
$ docker run --rm \
    -p 5432:5432 \
    -v my_app_data:/var/lib/postgresql/data \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    --name my-postgres-container postgres
```

To use it, first, change the name of the volume in the `docker-compose.yml` file.

Then, create the database container:

```text
$ docker-compose up
```

To stop it and remove the containers (but not the named volume(s) that contain the data) do:

```text
$ docker-compose down
```

(If you stop the containers by hitting `ctrl-c`, it won't remove the containers.)

To enter the database container, find its ID:

```text
$ docker container ls
```

Then, these commands:

```text
$ docker container exec -it <container_id> bash
# su -u postgres
$ psql
# \l
```

(The terminal prompt will change.)

You get information about the volumes like this:

```text
$ docker volume ls
$ docker volume inspect <volume_name>
```
