TWILIO_SRC_ID ?= 'default:twilio'

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
	TWILIO_SRC_ID=$(TWILIO_SRC_ID) pipenv run flask run -p 50010
