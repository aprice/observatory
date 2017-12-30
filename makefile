# Targets:
# 	all: Builds the code
#   setup: Install toolchain dependencies (e.g. gox)
#   lint: Run linters against source code
# 	format: Formats the source files
# 	build: Builds the code for target OS/arch combinations
# 	install: Installs the command to the GOPATH
# 	clean: cleans the build
# 	dep: Installs referenced projects
#	test: Runs the tests
#   package: Build, add aux files (eg README), zip/tar
#	debug: Show parameters.
#
# Parameters:
# 	VERSION: release version in semver format
#	BUILD_TAGS: additional build tags to pass to go build
#	DISTDIR: path to save distribution files
#	RPTDIR: path to save build/test reports


# Make is !@#$ing weird.
E :=
BSLASH := \$E
FSLASH := /

# Directories
WD := $(subst $(BSLASH),$(FSLASH),$(shell pwd))
MD := $(subst $(BSLASH),$(FSLASH),$(shell dirname "$(realpath $(lastword $(MAKEFILE_LIST)))"))
PKGDIR = $(MD)
CMDDIR = $(PKGDIR)/cmd
DISTDIR ?= $(WD)/dist
RPTDIR ?= $(WD)/reports
GP = $(subst $(BSLASH),$(FSLASH),$(GOPATH))

# Parameters
VERSION ?= $(shell git -C "$(MD)" describe --tags --dirty=-dev | tail -c +1)
COMMIT_ID := $(shell git -C "$(MD)" rev-parse --short=8 HEAD)
BUILD_TAGS ?= debug
PKG = github.com/aprice/observatory
CMDPKG = $(PKG)/cmd
CMDS := $(shell find "$(CMDDIR)/" -mindepth 1 -maxdepth 1 -type d | sed 's/ /\\ /g' | xargs -n1 basename)
NAME = coordinator
DOC = README.md LICENSE
BENCHCPUS ?= 1,2,4

# Commands
GOCMD = go
ARCHES = 386 amd64
OSES = windows linux darwin
OUTTPL = $(DISTDIR)/$(NAME)-$(VERSION)-{{.OS}}_{{.Arch}}/{{.Dir}}
LDFLAGS = -X $(PKG).Version=$(VERSION) -X $(PKG).Build=$(COMMIT_ID) -X $(DEP_SHARED).Version=$(VERSION) -X $(DEP_SHARED).Build=$(DEP_SHARED_COMMIT_ID)
GOBUILD = gox -rebuild -gocmd="$(GOCMD)" -arch="$(ARCHES)" -os="$(OSES)" -output="$(OUTTPL)" -tags "$(BUILD_TAGS)" -ldflags "$(LDFLAGS)"
GOCLEAN = $(GOCMD) clean
GOINSTALL = $(GOCMD) install -a -tags "$(BUILD_TAGS)" -ldflags "$(LDFLAGS)"
GOTEST = $(GOCMD) test -v -tags "$(BUILD_TAGS)"
GOCOVER = gocov test -tags "$(BUILD_TAGS)" $(PKG)/... | gocov-xml > "$(RPTDIR)/coverage.xml"
GOLINT = gometalinter --deadline=30s --tests --disable=aligncheck --disable=gocyclo --disable=gotype
GODEP = $(GOCMD) get -d -t
GOFMT = goreturns -w
GOCOV = $(GOCMD) tool cover
GOBENCH = $(GOCMD) test -v -tags "$(BUILD_TAGS)" -cpu=$(BENCHCPUS) -run=^$$ -bench=. -benchmem -outputdir "$(RPTDIR)"
GZCMD = tar -czf
ZIPCMD = zip -r
SHACMD = sha256sum
SLOCCMD = cloc --by-file --xml --exclude-dir="vendor" --include-lang="Go"
XUCMD = go2xunit

# Dynamic Targets
BUILD_TARGETS := $(addprefix build-,$(CMDS))
INSTALL_TARGETS := $(addprefix install-,$(CMDS))

.PHONY: setup setup-build setup-format setup-lint setup-reports clean format dep lint test test-mongo test-all bench install debug

all: debug setup dep format lint test-all bench build dist

setup: setup-dirs setup-build setup-format setup-lint setup-reports setup-cover

setup-reports: setup-dirs
	go get github.com/tebeka/go2xunit
	go get github.com/ryancox/gobench2plot
setup-build: setup-dirs
	go get github.com/mitchellh/gox
setup-format: setup-dirs
	go get github.com/sqs/goreturns
setup-lint: setup-dirs
	go get github.com/alecthomas/gometalinter
	gometalinter --install
setup-dirs:
	mkdir -p "$(RPTDIR)"
	mkdir -p "$(DISTDIR)"
setup-cover:
	go get github.com/axw/gocov/gocov
	go get github.com/AlekSi/gocov-xml
clean:
	$(GOCLEAN) $(PKG)
	rm -rf "$(DISTDIR)"/*
	rm -f "$(RPTDIR)"/*
format:
	$(GOFMT) "$(PKGDIR)"
dep:
	if [ ! -e "$(DEP_SHARED_DIR)" ]; then git clone "$(DEP_SHARED_REPO)" "$(DEP_SHARED_DIR)"; fi
	cd "$(DEP_SHARED_DIR)"; git pull
	$(GODEP) $(PKG)/...
	$(GODEP) $(DEP_SHARED)/...
lint: setup-dirs dep
	$(GOLINT) "$(PKGDIR)" | tee "$(RPTDIR)/lint.out"
test: setup-dirs clean dep
	$(GOTEST) $$(go list "$(PKG)/..." | grep -v /vendor/) | tee "$(RPTDIR)/test.out"
cover: setup-dirs setup-cover clean dep
	$(GOCOVER)
test-mongo: BUILD_TAGS += mongo
test-mongo: setup-dirs clean dep
	$(GOTEST) $$(go list "$(PKG)/..." | grep -v /vendor/) | tee "$(RPTDIR)/test.out"
test-all: BUILD_TAGS += mongo
test-all: setup-dirs clean dep
	$(GOTEST) $$(go list "$(PKG)/..." | grep -v /vendor/) | tee "$(RPTDIR)/test.out"
bench: BUILD_TAGS += mongo
bench: setup-dirs clean dep
	$(GOBENCH) $$(go list "$(PKG)/..." | grep -v /vendor/) | tee "$(RPTDIR)/bench.out" > goplot2xml > "$(RPTDIR)/bench.xml"
test-report: setup-dirs
	cd $(PKGDIR);$(SLOCCMD) --out="$(RPTDIR)/cloc.xml" . | tee "$(RPTDIR)/cloc.out"
	cat "$(RPTDIR)/test.out" | $(XUCMD) -output "$(RPTDIR)/tests.xml"
list-deps: setup-dirs
	rm -f "$(RPTDIR)/deps.out"
	go list -f '{{join .Deps "\n"}}' "$(CMDPKG)/coordinator" | sort | uniq | xargs -I {} sh -c "go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' {} | tee -a '$(RPTDIR)/deps.out'"
build: $(CMDS)
$(CMDS): setup-dirs dep
	$(GOBUILD) "$(CMDPKG)/$@" | tee "$(RPTDIR)/build-$@.out"
install: $(INSTALL_TARGETS)
$(INSTALL_TARGETS):
	$(GOINSTALL) "$(CMDPKG)/$(subst install-,,$@)"

dist: clean build
	for docfile in $(DOC); do \
		for dir in "$(DISTDIR)"/*; do \
			cp "$(PKGDIR)/$$docfile" "$$dir/"; \
		done; \
	done
	cd "$(DISTDIR)"; for dir in *linux*; do $(GZCMD) "$(basename "$$dir").tar.gz" "$$dir"; done
	cd "$(DISTDIR)"; for dir in ./*windows*; do $(ZIPCMD) "$(basename "$$dir").zip" "$$dir"; done
	cd "$(DISTDIR)"; for dir in *darwin*; do $(GZCMD) "$(basename "$$dir").tar.gz" "$$dir"; done
	cd "$(DISTDIR)"; find . -maxdepth 1 -type f -printf "$(SHACMD) %P | tee \"./%P.sha\"\n" | sh
	$(info "Built v$(VERSION), build $(COMMIT_ID)")
debug:
	$(info MD=$(MD))
	$(info WD=$(WD))
	$(info PKG=$(PKG))
	$(info PKGDIR=$(PKGDIR))
	$(info DISTDIR=$(DISTDIR))
	$(info VERSION=$(VERSION))
	$(info COMMIT_ID=$(COMMIT_ID))
	$(info BUILD_TAGS=$(BUILD_TAGS))
	$(info CMDS=$(CMDS))
	$(info BUILD_TARGETS=$(BUILD_TARGETS))
	$(info INSTALL_TARGETS=$(INSTALL_TARGETS))
