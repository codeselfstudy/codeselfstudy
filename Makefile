.PHONY: clean docs format dev production dockerdown

help:
	@echo "clean - remove junk files"
	@echo "docs - build the documentation and serve it in a browser"
	@echo "format - format the code"
	@echo "docker - build the docker containers and run them with Docker Compose"
	@echo "dockerdown - shut down the Docker containers"
	@echo "production - build and run the application for production with Docker Compose"

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
	docker-compose up --build -d -f docker-compose.production.yml

dev:
	docker-compose up --build -f docker-compose.dev.yml

dockerdown:
	docker-compose down
