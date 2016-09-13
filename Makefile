JUJU1_VERSIONS = 1.24.7 1.25.6
JUJU2_VERSIONS = 2.0-beta17
VERSIONS = $(JUJU1_VERSIONS) $(JUJU2_VERSIONS)
GO_VERSION = 1.6
TARBALLS = $(foreach version,$(VERSIONS),juju-core_$(version).tar.gz)
BUILT_VERSIONS = $(foreach version,$(VERSIONS),$(version)/$(version))
JUJU_TARBALL = juju-core_$(JUJU_VERSION).tar.gz
JUJU_PATCH = patches/juju-core_$(JUJU_VERSION).patch

.PHONY: build
build: $(BUILT_VERSIONS)

$(BUILT_VERSIONS):
	for VERSION in $(JUJU1_VERSIONS); do \
	    $(MAKE) build-common PATH=$$PATH JUJU_VERSION=$$VERSION; \
	done
	for VERSION in $(JUJU2_VERSIONS); do \
		$(MAKE) build-common PATH=/usr/lib/go-$(GO_VERSION)/bin:$$PATH JUJU_VERSION=$$VERSION; \
	done

.PHONY: ci-test
ci-test:
	sudo apt-get -y install --force-yes \
		wget \
		python3-pip \
		golang-go \
		golang-1.6
	python3 -m pip install jujuclient
	make test

.PHONY: build-common
build-common: $(JUJU_TARBALL) $(JUJU_PATCH)
	tar -C $(JUJU_VERSION) --strip=1 -z -xf $(JUJU_TARBALL)
	patch -p0 < $(JUJU_PATCH)
	cd $(JUJU_VERSION) && GOPATH=$(shell pwd)/$(JUJU_VERSION) PATH=$(PATH) go build

.PHONY: install
install: 
	for VERSION in $(VERSIONS); do \
	    $(MAKE) install-common JUJU_VERSION=$$VERSION; \
	done

.PHONY: install-common
install-common: $(JUJU_VERSION)/$(JUJU_VERSION)
	install -D $(JUJU_VERSION)/$(JUJU_VERSION) $(DESTDIR)/usr/bin/fake-juju-$(JUJU_VERSION)

.PHONY: clean
clean:
	for VERSION in $(VERSIONS) ; do \
	    JUJU_VERSION=$$VERSION make clean-common; \
	done	

.PHONY: clean-common
clean-common:
	rm -f $(JUJU_TARBALL)
	rm -rf $(JUJU_VERSION)/src
	rm -f $(JUJU_VERSION)/$(JUJU_VERSION)
	rm -rf _trial_temp
	rm -f tests/*.pyc

.PHONY: test
# Use xargs here so that we don't throw away the return codes, and correctly fail if any of the tests fail
test: $(BUILT_VERSIONS)
	@echo -n $(VERSIONS) | xargs -t -d' ' -I {} env JUJU_VERSION={} python3 -m unittest tests.test_fake

juju-core_%.tar.gz:
	#wget https://launchpad.net/juju-core/$(shell ((echo $* | grep -q beta) && echo trunk) || (echo $* | cut -f 1,2 -d .))/$*/+download/$@
	case $* in \
		1.*) \
			wget https://launchpad.net/juju-core/$(shell ((echo $* | grep -q beta) && echo trunk) || (echo $* | cut -f 1,2 -d .))/$*/+download/$@;; \
		2.*) \
	    	wget https://launchpad.net/juju/$(shell (echo $* | cut -f 1 -d - | cut -f 1,2 -d .))/$*/+download/$@;; \
	esac
