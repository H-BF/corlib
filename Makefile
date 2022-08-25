export GOSUMDB=off
export GO111MODULE=on
#export GOPROXY=https://goproxy.io,direct

$(value $(shell [ ! -d "$(CURDIR)/bin" ] && mkdir -p "$(CURDIR)/bin"))
GOBIN:=$(CURDIR)/bin
GOLANGCI_BIN:=$(GOBIN)/golangci-lint
GOLANGCI_REPO:=https://github.com/golangci/golangci-lint
GOLANGCI_LATEST_VERSION?= $(shell git ls-remote --tags --refs --sort='v:refname' $(GOLANGCI_REPO)|tail -1|egrep -o "v[0-9]+.*")

ifneq ($(wildcard $(GOLANGCI_BIN)),)
	GOLANGCI_CUR_VERSION:=v$(shell $(GOLANGCI_BIN) --version|sed -E 's/.* version (.*) built from .* on .*/\1/g')
else
	GOLANGCI_CUR_VERSION:=
endif

# install linter tool
.PHONY: install-linter
install-linter:
	$(info GOLANGCI-LATEST-VERSION=$(GOLANGCI_LATEST_VERSION))
ifeq ($(filter $(GOLANGCI_CUR_VERSION), $(GOLANGCI_LATEST_VERSION)),)
	$(info Installing GOLANGCI-LINT $(GOLANGCI_LATEST_VERSION)...)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LATEST_VERSION)
	@chmod +x $(GOLANGCI_BIN)
else
	@echo "GOLANGCI-LINT is need not install"
endif

# run full lint like in pipeline
.PHONY: lint
lint: install-linter
	$(info GOBIN=$(GOBIN))
	$(info GOLANGCI_BIN=$(GOLANGCI_BIN))
	$(GOLANGCI_BIN) cache clean && \
	$(GOLANGCI_BIN) run --config=$(CURDIR)/.golangci.yaml -v $(CURDIR)/...

# install project dependencies
.PHONY: go-deps
go-deps:
	$(info Install dependencies...)
	@go mod tidy && go mod vendor && go mod verify

.PHONY: test
test:
	$(info Running tests...)
	go clean -testcache && go test -v  ./...




