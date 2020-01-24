# Coding Puzzles

A tool to scrape and serve coding puzzles.

## Technologies

"MERN stack"

- Express.js for the API
- Redis for session storage in Express.js
- MongoDB for persistent storage
- React for an SPA (with Bootstrap)
- Docker for easy development and deployment
- Python 3 scraping tools (possibly to be replaced with puppeteer scrapers)

![General overview](./misc/servers.png)

## Development

To run the application:

- Be sure that you have Docker installed.
- Turn off your local MongoDB server, because this will run it's own Mongo in Docker on the default port (27017).
- Start the application with Docker Compose:

```text
$ docker-compose up --build
```

Then visit [localhost:4444](http://localhost:4444/).

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

### How It Works

See the `docker-compose.yml` file for an overview of the services that are running.

- Nginx serves the app in development and production, proxying to the other servers behind it. It serves on port 80, but in development, you can visit it at port 4444.
- There is a React SPA being served in another container with another instance of nginx.
- Express.js is running on port 5000, mounted at `/api`. (`/api` is not visible in the Express code, so that might be confusing if you don't know to expect it. See the nginx configuration files.)
- A Redis container will handle Express sessions (after users authenticate with the forum).
- A MongoDB container will store the coding puzzle and user data. (This is an experiment with "MERN stack".)

