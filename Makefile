JUJUCLIENT_DOWNLOADS = $(shell pwd)/tests/jujuclient-archive
JUJUCLIENT_REQ = $(JUJUCLIENT_DOWNLOADS)/requirements


ifdef JUJU_VERSION  #############################
# for one version

GO_VERSION = 1.6
JUJU_TARBALL = juju-core_$(JUJU_VERSION).tar.gz
JUJU_PATCH = patches/juju-core_$(JUJU_VERSION).patch
INSTALLDIR = $(DESTDIR)/usr/bin
INSTALLED = $(INSTALLDIR)/fake-juju-$(JUJU_VERSION)

$(JUJU_VERSION)/$(JUJU_VERSION): $(JUJU_VERSION)/fake-juju.go
	case $(JUJU_VERSION) in \
		1.*) $(MAKE) build-common PATH=$$PATH JUJU_VERSION=$(JUJU_VERSION) ;;\
		2.*) $(MAKE) build-common PATH=/usr/lib/go-$(GO_VERSION)/bin:$$PATH JUJU_VERSION=$(JUJU_VERSION) ;;\
	esac

juju-core_%.tar.gz:
	case $* in \
		1.*) PROJECT="juju-core" ;;\
		2.*) PROJECT="juju" ;;\
	esac;\
	wget https://launchpad.net/$$PROJECT/$(shell (echo $* | cut -f 1 -d - | cut -f 1,2 -d .))/$*/+download/$@

.PHONY: build
build: $(JUJU_VERSION)/$(JUJU_VERSION)

.PHONY: build-common
build-common: $(JUJU_TARBALL) $(JUJU_PATCH)
	rm -rf $(JUJU_VERSION)/src  # Go doesn't play nice with existing files.
	tar -C $(JUJU_VERSION) --strip=1 -z -xf $(JUJU_TARBALL)
	patch -p0 < $(JUJU_PATCH)
	cd $(JUJU_VERSION) && GOPATH=$(shell pwd)/$(JUJU_VERSION) PATH=$(PATH) go build

.PHONY: install
install: $(JUJU_VERSION)/$(JUJU_VERSION)
	install -D $(JUJU_VERSION)/$(JUJU_VERSION) $(INSTALLED)

.PHONY: install-dev
install-dev: $(JUJU_VERSION)/$(JUJU_VERSION)
	mkdir -p $(INSTALLDIR)
	ln -s --backup=existing --suffix .orig $(shell pwd)/$(JUJU_VERSION)/$(JUJU_VERSION) $(INSTALLED)

.PHONY: clean
clean:
	rm -f $(JUJU_TARBALL)
	rm -rf $(JUJU_VERSION)/src
	rm -f $(JUJU_VERSION)/$(JUJU_VERSION)
	rm -rf _trial_temp
	rm -f tests/*.pyc

.PHONY: test
test: $(JUJU_VERSION)/$(JUJU_VERSION)
	env JUJU_VERSION=$(JUJU_VERSION) python3 -m unittest tests.test_fake


else  ###########################################
# for all versions

JUJU1_VERSIONS = 1.24.7 1.25.6
JUJU2_VERSIONS = 2.0.0
VERSIONS = $(JUJU1_VERSIONS) $(JUJU2_VERSIONS)
BUILT_VERSIONS = $(foreach version,$(VERSIONS),$(version)/$(version))

$(BUILT_VERSIONS):
	for VERSION in $(VERSIONS); do \
	    $(MAKE) build JUJU_VERSION=$$VERSION; \
	done

.PHONY: build
build: $(BUILT_VERSIONS)

.PHONY: install
install:
	for VERSION in $(VERSIONS); do \
	    $(MAKE) install JUJU_VERSION=$$VERSION; \
	done

.PHONY: install-dev
install-dev:
	for VERSION in $(VERSIONS); do \
	    $(MAKE) install-dev JUJU_VERSION=$$VERSION; \
	done

.PHONY: clean
clean:
	for VERSION in $(VERSIONS) ; do \
		$(MAKE) clean JUJU_VERSION=$$VERSION; \
	done	

.PHONY: test
# Use xargs here so that we don't throw away the return codes, and correctly fail if any of the tests fail
test: $(BUILT_VERSIONS)
	#@echo -n $(VERSIONS) | xargs -t -d' ' -I {} env JUJU_VERSION={} python3 -m unittest tests.test_fake
	@echo -n $(VERSIONS) | xargs -t -d' ' -I {} $(MAKE) test JUJU_VERSION={}


endif  ##########################################

.PHONY: ci-test
ci-test:
	sudo apt-get -y install --force-yes \
		wget \
		python3-pip \
		golang-go \
		golang-1.6
	# See tests/jujuclient-archive/UPGRADE when a newer jujuclient version is needed.
	sudo python3 -m pip install \
		--ignore-installed \
        --no-cache-dir \
		--no-index \
		--find-links $(JUJUCLIENT_DOWNLOADS) \
		--requirement $(JUJUCLIENT_REQ)
	$(MAKE) test
