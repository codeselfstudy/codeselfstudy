.PHONY: clean docs format initialize

help:
	@echo "clean - remove junk files"
	@echo "docs - build the documentation and serve it in a browser"
	@echo "format - format the code"
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

initialize:
	echo 'not implemented'
# 	cd containers/gatsby/src/content
# 	git submodule init
# 	git submodule update
# 	./scripts/check_env.sh
