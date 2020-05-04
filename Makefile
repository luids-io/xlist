# Makefile for building xlist

# Project binaries
COMMANDS=xlistd xlistc xlget
BINARIES=$(addprefix bin/,$(COMMANDS))

# Used to populate version in binaries
VERSION=$(shell git describe --match 'v[0-9]*' --dirty='.m' --always)
REVISION=$(shell git rev-parse HEAD)$(shell if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi)
DATEBUILD=$(shell date +%FT%T%z)

# Compilation opts
GOPATH?=$(HOME)/go
SYSTEM:=
CGO_ENABLED:=0
BUILDOPTS:=-v
BUILDLDFLAGS=-ldflags '-s -w -X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.Build=$(DATEBUILD) $(EXTRA_LDFLAGS)'

# Print output
WHALE = "+"

.PHONY: all binaries database clean docker
all: binaries


FORCE:


# Build a binary from a cmd.
bin/%: cmd/% FORCE
	@echo "$(WHALE) $@${BINARY_SUFFIX}"
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(SYSTEM) \
		go build $(BUILDOPTS) -o $@${BINARY_SUFFIX} ${BUILDLDFLAGS} ./$< 


binaries: $(BINARIES)
	@echo "$(WHALE) $@"

database:
	@echo "$(WHALE) $@"
	database/scripts/update_csv.sh
	database/scripts/update_readme.sh

clean:
	@echo "$(WHALE) $@"
	@rm -f $(BINARIES)
	@rmdir bin

docker:
	@echo "$(WHALE) $@"
	docker build -t xlistd -f Dockerfile.xlistd .
	docker build -t xlget -f Dockerfile.xlget .

## Targets for Makefile.release
.PHONY: release
release:
	@$(if $(value BINARY),, $(error Undefined BINARY))
	@$(if $(value COMMAND),, $(error Undefined COMMAND))
	@echo "$(WHALE) $@"
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(SYSTEM) \
		    go build $(BUILDOPTS) ${BUILDLDFLAGS} -o $(BINARY) ./cmd/$(COMMAND)

# tests
.PHONY: test test-core test-plugin
test: test-core test-plugin
	@echo "$(WHALE) $@"


test-core:
	@echo "$(WHALE) $@"
	( cd pkg/listbuilder ; GO111MODULE=on go test -v -race ./...)


test-plugin:
	@echo "$(WHALE) $@"
	( cd pkg/components ; GO111MODULE=on go test -v -race ./... )
	( cd pkg/wrappers ; GO111MODULE=on go test -v -race ./... )

