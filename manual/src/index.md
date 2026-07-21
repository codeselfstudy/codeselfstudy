# Code Self Study Manual

Welcome to the Code Self Study website codebase. This mdBook manual is a short guide for getting the app running and understanding how the project is organized.

The repo is a Bun workspace with an Astro frontend (`apps/web/`) and a Go + Echo backend (`apps/api/`). In production a single Go binary serves the prerendered site, the JSON API, and a future WebSocket endpoint — there is no JavaScript runtime on the server.

## Project Goals

- Help members find something in common to work on.
- Support meetup activities and discussion.

Current ideas include authentication, daily puzzles, marking completed puzzles, and surfacing tasks by difficulty.

## Start Here

- Read "Getting Started" to set up your environment and run the app.
- Skim "Architecture" for the high-level shape of the two apps and how they fit together.
- Use "Project Structure" to find features and data quickly.
- Check "Routing & Data" for Astro routing conventions.
