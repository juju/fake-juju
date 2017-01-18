JUJU_VERSIONS = 1.25.6 2.0.0 2.0.2

JUJUCLIENT_DOWNLOADS = $(shell pwd)/tests/jujuclient-archive
JUJUCLIENT_REQ = $(JUJUCLIENT_DOWNLOADS)/requirements

PYTHON = python
TOX ?= tox

.PHONY: all
all: build

.PHONY: build
build:
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) -C $$VERSION build; \
	done

.PHONY: install
install:
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) -C $$VERSION install; \
	done

.PHONY: clean
clean:
	for VERSION in $(JUJU_VERSIONS) ; do \
	    $(MAKE) -C $$VERSION clean; \
	done

.PHONY: test
test: build
	$(MAKE) go-test
	$(MAKE) py-test

.PHONY: go-test
go-test:
	# Use xargs here so that we don't throw away the return codes, and
	# correctly fail if any of the tests fail
	@echo -n $(JUJU_VERSIONS) | xargs -t -d' ' -I {} $(MAKE) -C {} test

.PHONY: py-test
py-test:
	cd python && $(TOX)
