VERSIONS = 1.22.6 1.24.6 1.24.7 1.25.0 1.25.3
TARBALLS = $(foreach version,$(VERSIONS),juju-core_$(version).tar.gz)
BUILT_VERSIONS = $(foreach version,$(VERSIONS),$(version)/$(version))
JUJU_TARBALL = juju-core_$(JUJU_VERSION).tar.gz
JUJU_PATCH = patches/juju-core_$(JUJU_VERSION).patch

build: $(BUILT_VERSIONS)
.PHONY: build

$(BUILT_VERSIONS):
	for VERSION in $(VERSIONS); do \
	    $(MAKE) build-common JUJU_VERSION=$$VERSION; \
	done

build-common: $(JUJU_TARBALL) $(JUJU_PATCH)
	tar -C $(JUJU_VERSION) --strip=1 -z -xf $(JUJU_TARBALL)
	patch -p0 < $(JUJU_PATCH)
	cd $(JUJU_VERSION) && GOPATH=$(shell pwd)/$(JUJU_VERSION) go build
.PHONY: build-common

install: 
	for VERSION in $(VERSIONS) ; do \
	    $(MAKE) install-common JUJU_VERSION=$$VERSION; \
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
	rm -f $(JUJU_TARBALL)
	rm -rf $(JUJU_VERSION)/src
	rm -f $(JUJU_VERSION)/$(JUJU_VERSION)
	rm -rf _trial_temp
	rm -f tests/*.pyc
.PHONY: clean-common


# Use xargs here so that we don't throw away the return codes, and correctly fail if any of the tests fail
test: $(BUILT_VERSIONS)
	@echo -n $(VERSIONS) | xargs -t -d' ' -I {} env JUJU_VERSION={} python -m unittest tests.test_fake
.PHONY: test

juju-core_%.tar.gz:
	wget https://launchpad.net/juju-core/$(shell echo $* | cut -f 1,2 -d .)/$*/+download/$@
