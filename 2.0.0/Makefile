GO_VERSION = 1.6
GO ?= /usr/lib/go-$(GO_VERSION)/bin/go
GO_PATH = $(CURDIR)

JUJU_VERSION = $(shell basename $(CURDIR))
JUJU_PATCH = juju-core.patch
JUJU_INSTALLDIR = $(DESTDIR)/usr/bin
JUJU_INSTALLED = $(JUJU_INSTALLDIR)/fake-juju-$(JUJU_VERSION)

build:
	rm -rf src  # Go doesn't play nice with existing files.
	$(MAKE) src
	patch -p0 < $(JUJU_PATCH)
	GOPATH=$(GO_PATH) $(GO) build -v

install:
	install -D $(JUJU_VERSION) $(JUJU_INSTALLED)

%: force
	@$(MAKE) -f ../common/makefile.mk $@
force: ;
