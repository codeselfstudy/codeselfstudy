# Coding Puzzles

A tool to scrape and serve coding puzzles.

## Technologies

- Python 3 scraping tools
- Express.js for the API
- MongoDB for persistent storage (many people in the group know it)

## Development

To run the application:

- Be sure that you have Docker installed.
- Turn off your local MongoDB server, because this will run it's own Mongo in Docker on the default port (27017).
- Start the application with Docker Compose:

```text
$ docker-compose up --build
```

Then visit [localhost:3000](http://localhost:3000/).

See the routes files for the available routes.

To shutdown the containers:

```text
$ docker-compose down
```

(If you stop the containers by hitting `ctrl-c`, it won't remove the containers.)

To enter a running database container, find its ID:

```text
$ docker container ps
```

Then, these commands:

```text
$ docker container exec -it <container_id> sh
```

You get information about the volumes like this:

```text
$ docker volume ls
$ docker volume inspect <volume_name>
```
