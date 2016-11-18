
ifdef JUJU_VERSIONS
# If someone explicitly set JUJU_VERSIONS then we ignore JUJU_VERSION.
JUJU_VERSION =
endif


ifdef JUJU_VERSION  #############################
# for one version

GO_VERSION = 1.6
JUJU_TARBALL = juju-core_$(JUJU_VERSION).tar.gz
JUJU_PATCH = patches/juju-core_$(JUJU_VERSION).patch
GO_PATH = $(JUJU_VERSION)
JUJU_SRC = $(GO_PATH)/src
JUJU_UNPACKED_CLEAN = $(JUJU_VERSION)/.unpacked-clean
INSTALLDIR = $(DESTDIR)/usr/bin
INSTALLED = $(INSTALLDIR)/fake-juju-$(JUJU_VERSION)

$(JUJU_VERSION)/$(JUJU_VERSION): $(JUJU_VERSION)/fake-juju.go $(JUJU_PATCH)
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

$(JUJU_UNPACKED_CLEAN): $(JUJU_TARBALL)
	mkdir -p $(JUJU_UNPACKED_CLEAN)
	tar -C $(JUJU_UNPACKED_CLEAN) --strip=1 -z -xf $(JUJU_TARBALL)

.PHONY: update-patch
update-patch: $(JUJU_SRC) $(JUJU_UNPACKED_CLEAN)
	diff -U 3 -r --no-dereference ./$(JUJU_UNPACKED_CLEAN)/src ./$(JUJU_SRC) > $(JUJU_PATCH); \
	echo " -- diff exited with $$? --"
	rm -rf $(JUJU_UNPACKED_CLEAN)

.PHONY: build
build: $(JUJU_VERSION)/$(JUJU_VERSION) py-build

.PHONY: build-common
build-common: $(JUJU_TARBALL) $(JUJU_PATCH)
	rm -rf $(JUJU_SRC)  # Go doesn't play nice with existing files.
	tar -C $(JUJU_VERSION) --strip=1 -z -xf $(JUJU_TARBALL)
	patch -p0 < $(JUJU_PATCH)
	cd $(JUJU_VERSION) && GOPATH=$(shell pwd)/$(GO_PATH) PATH=$(PATH) go build

.PHONY: install
install: $(JUJU_VERSION)/$(JUJU_VERSION) py-install-dev
	install -D $(JUJU_VERSION)/$(JUJU_VERSION) $(INSTALLED)

.PHONY: install-dev
install-dev: $(JUJU_VERSION)/$(JUJU_VERSION) py-install-dev
	mkdir -p $(INSTALLDIR)
	ln -s --backup=existing --suffix .orig $(shell pwd)/$(JUJU_VERSION)/$(JUJU_VERSION) $(INSTALLED)

.PHONY: clean
clean: py-clean
	rm -f $(JUJU_TARBALL)
	rm -rf $(JUJU_SRC)
	rm -f $(JUJU_VERSION)/$(JUJU_VERSION)
	rm -rf _trial_temp
	rm -f tests/*.pyc

.PHONY: test
test: $(JUJU_VERSION)/$(JUJU_VERSION) py-test
	env JUJU_VERSION=$(JUJU_VERSION) python3 -m unittest tests.test_fake


else  ###########################################
# for all versions

ifndef JUJU_VERSIONS
JUJU1_VERSIONS = 1.25.6
JUJU2_VERSIONS = 2.0.0
JUJU_VERSIONS = $(JUJU1_VERSIONS) $(JUJU2_VERSIONS)
endif
BUILT_VERSIONS = $(foreach version,$(JUJU_VERSIONS),$(version)/$(version))

$(BUILT_VERSIONS):
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) build JUJU_VERSION=$$VERSION SKIP_PYTHON_LIB=TRUE; \
	done

.PHONY: update-patch
update-patch:
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) update-patch JUJU_VERSION=$$VERSION; \
	done

.PHONY: build
build: $(BUILT_VERSIONS) py-build

.PHONY: install
install: py-install
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) install JUJU_VERSION=$$VERSION SKIP_PYTHON_LIB=TRUE; \
	done

.PHONY: install-dev
install-dev: py-install-dev
	for VERSION in $(JUJU_VERSIONS); do \
	    $(MAKE) install-dev JUJU_VERSION=$$VERSION SKIP_PYTHON_LIB=TRUE; \
	done

.PHONY: clean
clean: py-clean
	for VERSION in $(JUJU_VERSIONS) ; do \
		$(MAKE) clean JUJU_VERSION=$$VERSION SKIP_PYTHON_LIB=TRUE; \
	done	

.PHONY: test
# Use xargs here so that we don't throw away the return codes, and correctly fail if any of the tests fail
test: $(BUILT_VERSIONS) py-test
	#@echo -n $(JUJU_VERSIONS) | xargs -t -d' ' -I {} env JUJU_VERSION={} python3 -m unittest tests.test_fake
	@echo -n $(JUJU_VERSIONS) | xargs -t -d' ' -I {} $(MAKE) test JUJU_VERSION={} SKIP_PYTHON_LIB=TRUE


endif  ##########################################

# for the Python library

PYTHON = python
ifndef PYTHON_INSTALLDIR
PYTHON_INSTALLDIR = $(DESTDIR)/usr/lib/python2.7/dist-packages
endif
ifndef PYTHON_SCRIPTDIR
PYTHON_SCRIPTDIR = $(DESTDIR)/usr/local/bin
endif
MODULEDIR = $(DESTDIR)/usr/local/bin
PYTHON_INSTALL_OPTION = --install-lib $(PYTHON_INSTALLDIR) --install-scripts $(PYTHON_SCRIPTDIR)

PYTHON_LIB_ROOT = $(shell pwd)/python
# TODO: read from python/fakejuju/__init__.py
PYTHON_LIB_VERSION = 0.9.0b1
PYTHON_LIB_SOURCE_TARBALL = $(PYTHON_LIB_ROOT)/dist/fakejuju-$(PYTHON_LIB_VERSION).tar.gz

$(PYTHON_LIB_SOURCE_TARBALL): python/fakejuju
	echo $(PYTHON_LIB_SOURCE_TARBALL)
	cd python; \
	$(PYTHON) setup.py sdist

.PHONY: py-build
py-build: $(PYTHON_LIB_SOURCE_TARBALL)

.PHONY: py-install
py-install: py-build
	if [ ! "$(SKIP_PYTHON_LIB)" ]; then \
		mkdir -p $(PYTHON_INSTALLDIR); \
		cd python; \
		PYTHONPATH=$(PYTHON_INSTALLDIR):$$PYTHONPATH $(PYTHON) setup.py install $(PYTHON_INSTALL_OPTION); \
	fi

.PHONY: py-install-dev
py-install-dev:
	if [ ! "$(SKIP_PYTHON_LIB)" ]; then \
		ln -snv --backup=existing --suffix .orig $(PYTHON_LIB_ROOT)/fakejuju $(PYTHON_INSTALLDIR)/fakejuju; \
	fi

.PHONY: py-clean
py-clean:
	if [ ! "$(SKIP_PYTHON_LIB)" ]; then \
		rm -rf python/dist; \
		rm -rf python/fakejuju.egg-info; \
		rm -f python/fakejuju/*.pyc; \
		rm -f python/fakejuju/tests/*.pyc; \
	fi

.PHONY: py-test
py-test:
	if [ ! "$(SKIP_PYTHON_LIB)" ]; then \
		$(PYTHON) -m unittest discover -t $(PYTHON_LIB_ROOT) -s $(PYTHON_LIB_ROOT)/fakejuju; \
	fi


#################################################
# other targets

JUJUCLIENT_DOWNLOADS = $(shell pwd)/tests/jujuclient-archive
JUJUCLIENT_REQ = $(JUJUCLIENT_DOWNLOADS)/requirements

.PHONY: ci-test
ci-test:
	sudo add-apt-repository -y --update ppa:landscape/txjuju-daily
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
