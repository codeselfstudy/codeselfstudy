# Code Self Study Website

This is the new [Code Self Study](https://codeselfstudy.com/) website.

This is an unfinished work in progress. :construction: Attend a meetup to find out how to contribute.

## Goals:

- help people in the group find something in common to work on
- meetup activity
- learn Docker better

### Ideas so far:

- authenticate with the forum
- send a coding puzzle of the day into the forum, slack, and/or the browser extension so that interested people have a common task to discuss
- add new puzzles (links to other sites or original puzzles)
- mark puzzles that you've completed
- saving puzzles to do later
- use the browser extension to add new puzzles to the database?
- fetch puzzle by difficulty and type of problem
- commenting? (forum integration or separate)
- voting

See the [feature board](https://github.com/codeselfstudy/codeselfstudy/projects/1) for more ideas (and add your own).

## Documentation

For the documentation, see the [Code Self Study Wiki](https://github.com/codeselfstudy/codeselfstudy_wiki) (also a work in progress).

## Technologies

See the `docker-compose.*.yml` files for an overview of the services that are running.

- Gatsby + markdown files for pages and blog posts
- Express.js for the API ("MERN Stack")
- Redis for session storage in Express.js
- MongoDB for persistent storage
- Docker for easy development and deployment
- Web scraping tools in Elixir
- Coming soon: algorithm & data structure visualzation in Rust/WebAssembly.
- Coming soon: a Phoenix/Postgres server

### How It Works

- Nginx serves the app in development and production, proxying to the other servers behind it. It serves on port 80, but in development, you can visit it at port 4444.
- Gatsby serves the main pages (backed by a headless CMS)
- Express.js is running on port 5000, mounted at `/api`. (`/api` is not visible in the Express code, so that might be confusing if you don't know to expect it. See the nginx configuration files.)
- A Redis container will handle Express sessions (after users authenticate with [the forum](https://forum.codeselfstudy.com/)).
- A MongoDB container will store the coding puzzle and user data. (This is an experiment with "MERN stack".)

![General overview](https://wiki.codeselfstudy.com/images/servers.png)

## Development

To run the application:

- Load the content submodule
    - `$ cd containers/gatsby/src/content`
    - `$ git submodule init`
    - `$ git submodule update`
- Be sure that you have Docker installed.
- Start the application with Docker Compose. (Look inside the `Makefile` for commands to boot the app, like `make dev`.)

```text
$ docker-compose -f docker-compose.dev.yml up --build
```

(There is a separate Docker Compose file for production: `docker-compose.production.yml`.)

Then visit [localhost:4444](http://localhost:4444/) in development mode.

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

## LICENSES

The code is licensed under a BSD licensed (see the `./LICENSES` directory). Other sections of the code may have their own licenses. (Search for files named `LICENSE` if you want to generate a list without looking in the directories.)
