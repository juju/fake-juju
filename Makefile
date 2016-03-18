VERSIONS = 1.22.6 1.24.6 1.24.7 1.25.0 1.25.3
TARBALLS = $(foreach version,$(VERSIONS),juju-core_$(version).tar.gz)
BUILT_VERSIONS = $(foreach version,$(VERSIONS),$(version)/$(version))

build: $(BUILT_VERSIONS)
.PHONY: build

$(BUILT_VERSIONS):
	for VERSION in $(VERSIONS); do \
	    JUJU_VERSION=$$VERSION $(MAKE) build-common; \
	done

build-common: juju-core_$(JUJU_VERSION).tar.gz
	tar -C $(JUJU_VERSION) --strip=1 -z -xf juju-core_$(JUJU_VERSION).tar.gz
	patch -p0 < patches/juju-core_$(JUJU_VERSION).patch
	cd $(JUJU_VERSION) && GOPATH=$(shell pwd)/$(JUJU_VERSION) go build
.PHONY: build-common

install: 
	for VERSION in $(VERSIONS) ; do \
	    JUJU_VERSION=$$VERSION make install-common; \
	done
.PHONY: install

install-common: $(JUJU_VERSION)/$(JUJU_VERSION)
	install -D $(JUJU_VERSION)/$(JUJU_VERSION) $(DESTDIR)/usr/bin/fake-juju-$(JUJU_VERSION)
.PHONY: install-common

clean:
	for VERSION in $(VERSIONS) ; do \
	    JUJU_VERSION=$$VERSION make clean-common; \
	done	
.PHONY: clean

clean-common:
	rm -f juju-core_$(JUJU_VERSION).tar.gz
	rm -rf $(JUJU_VERSION)/src
	rm -f $(JUJU_VERSION)/$(JUJU_VERSION)
	rm -rf _trial_temp
	rm -f tests/*.pyc
.PHONY: clean-common


test:
	for VERSION in $(VERSIONS) ; do \
	    JUJU_VERSION=$$VERSION $(MAKE) test-common; \
	done
.PHONY: test

test-common: $(JUJU_VERSION)/$(JUJU_VERSION)
	JUJU_VERSION=$(JUJU_VERSION) python -m unittest tests.test_fake
.PHONY: test-common

juju-core_%.tar.gz:
	wget https://launchpad.net/juju-core/$(shell echo $* | cut -f 1,2 -d .)/$*/+download/$@
