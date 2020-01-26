.PHONY: clean docs

help:
	@echo "clean - remove junk files"
	@echo "docs - build the documentation and serve it in a browser"

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
	pushd scrapers && mix format && popd
