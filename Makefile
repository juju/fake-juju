VERSIONS = 1.22.6 1.24.6
TARBALLS = $(foreach version,$(VERSIONS),juju-core_$(version).tar.gz)

build: $(TARBALLS)
	for VERSION in $(VERSIONS) ; do \
	    JUJU_VERSION=$$VERSION make build-common; \
	done

build-common:
	tar -C $(JUJU_VERSION) --strip=1 -z -xf juju-core_$(JUJU_VERSION).tar.gz
	patch -p0 < patches/juju-core_$(JUJU_VERSION).patch
	cd $(JUJU_VERSION) && GOPATH=$(shell pwd)/$(JUJU_VERSION) go build

install: 
	for VERSION in $(VERSIONS) ; do \
	    JUJU_VERSION=$$VERSION make install-common; \
	done

install-common:
	mkdir -p $(DESTDIR)/usr/bin
	cp $(JUJU_VERSION)/$(JUJU_VERSION) $(DESTDIR)/usr/bin/fake-juju-$(JUJU_VERSION)

clean:
	for VERSION in $(VERSIONS) ; do \
	    JUJU_VERSION=$$VERSION make clean-common; \
	done	

clean-common:
	rm -f juju-core_$(JUJU_VERSION).tar.gz
	rm -rf $(JUJU_VERSION)/src
	rm -f $(JUJU_VERSION)/$(JUJU_VERSION)
	rm -rf _trial_temp
	rm -f tests/*.pyc


test:
	for VERSION in $(VERSIONS) ; do \
	    JUJU_VERSION=$$VERSION make test-common; \
	done	

test-common:
	JUJU_VERSION=$(JUJU_VERSION) python -m unittest tests.test_fake

juju-core_%.tar.gz:
	wget https://launchpad.net/juju-core/$(shell echo $* | cut -f 1,2 -d .)/$*/+download/$@
