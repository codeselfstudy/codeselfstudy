.PHONY: clean docs format dev production dockerdown initialize

help:
	@echo "clean - remove junk files"
	@echo "docs - build the documentation and serve it in a browser"
	@echo "format - format the code"
	@echo "dev - build the docker containers and run them with Docker Compose"
	@echo "dockerdown - shut down the Docker containers"
	@echo "production - build and run the application for production with Docker Compose"
	@echo "initialize - initial build steps to run after cloning the repo"

clean:
	find . -name '*.pyc' -exec rm -f {} +
	find . -name '*.pyo' -exec rm -f {} +
	find . -name '*~' -exec rm -f {} +
	find . -name '__pycache__' -exec rm -rf {} +

docs:
	echo "INSTRUCTIONS: point your browser at http://localhost:3000/"
	cd docs && mdbook serve

format:
	prettier **/*.{js,css,html,json} --write
	cd scrapers && mix format

production:
	docker-compose -f docker-compose.production.yml up --build -d

dev:
	./scripts/check_env.sh
	docker-compose -f docker-compose.dev.yml up --build

# This works in development
dockerdown:
	./scripts/check_env.sh
	docker-compose -f docker-compose.dev.yml down

initialize:
	cd containers/gatsby/src/content
	git submodule init
	git submodule update
	./scripts/check_env.sh
