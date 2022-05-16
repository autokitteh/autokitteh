.PHONY: all
all: pipenv mypy

.PHONY: pipenv
pipenv:
	pipenv install

.PHONY: mypy
mypy:
	pipenv run mypy --strict .

.PHONY: clean
clean:
	pipenv clean

.PHONY: run
run:
	pipenv run python3 -m autokitteh.pluginsd --debug twilioplugin
