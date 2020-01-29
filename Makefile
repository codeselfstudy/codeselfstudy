.PHONY: clean docs format docker dockerdown

help:
	@echo "clean - remove junk files"
	@echo "docs - build the documentation and serve it in a browser"
	@echo "format - format the code"
	@echo "docker - build the docker containers and run them with Docker Compose"
	@echo "dockerdown - shut down the Docker containers"

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

docker:
	docker-compose up --build

dockerdown:
	docker-compose down
