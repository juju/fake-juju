JUJU_VERSIONS = 1.25.6 2.0.0

JUJUCLIENT_DOWNLOADS = $(shell pwd)/tests/jujuclient-archive
JUJUCLIENT_REQ = $(JUJUCLIENT_DOWNLOADS)/requirements

PYTHON = python

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

.PHONY: test py-test
test: build
	# Use xargs here so that we don't throw away the return codes, and
	# correctly fail if any of the tests fail
	@echo -n $(JUJU_VERSIONS) | xargs -t -d' ' -I {} $(MAKE) -C {} test JUJU_VERSION={}

.PHONY: py-test
py-test:
	$(PYTHON) -m unittest discover -t python -s python/fakejuju

.PHONY: ci-test
ci-test:
	sudo apt-get -y install --force-yes \
		wget \
		python3-pip \
		golang-go \
		golang-1.6 \
		python-txjuju
	# See tests/jujuclient-archive/UPGRADE when a newer jujuclient version is needed.
	sudo python3 -m pip install \
		--ignore-installed \
		--no-cache-dir \
		--no-index \
		--find-links $(JUJUCLIENT_DOWNLOADS) \
		--requirement $(JUJUCLIENT_REQ)
	$(MAKE) test
