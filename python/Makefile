PYTHON = python

.PHONY: test
test:
	$(PYTHON) -m unittest discover -t $(shell pwd) -s $(shell pwd)/fakejuju

.PHONY: install-dev
install-dev:
	ln -s $(shell pwd)/fakejuju /usr/local/lib/python2.7/dist-packages/fakejuju
