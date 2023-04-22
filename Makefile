export GOSUMDB=off
export GO111MODULE=on
#export GOPROXY=https://goproxy.io,direct

$(value $(shell [ ! -d "$(CURDIR)/bin" ] && mkdir -p "$(CURDIR)/bin"))
GOBIN:=$(CURDIR)/bin
GOLANGCI_BIN:=$(GOBIN)/golangci-lint
GOLANGCI_REPO:=https://github.com/golangci/golangci-lint
GOLANGCI_LATEST_VERSION?= $(shell git ls-remote --tags --refs --sort='v:refname' $(GOLANGCI_REPO)|tail -1|egrep -o "v[0-9]+.*")

ifneq ($(wildcard $(GOLANGCI_BIN)),)
	GOLANGCI_CUR_VERSION:=v$(shell $(GOLANGCI_BIN) --version|sed -E 's/.*version (.*) built.*/\1/g')
else
	GOLANGCI_CUR_VERSION:=
endif

# install linter tool
.PHONY: install-linter
install-linter: ##install linter tool
ifeq ($(filter $(GOLANGCI_CUR_VERSION), $(GOLANGCI_LATEST_VERSION)),)
	$(info Cur linter varsion '$(GOLANGCI_CUR_VERSION)' - installing  '$(GOLANGCI_LATEST_VERSION)'...)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LATEST_VERSION)
	@chmod +x $(GOLANGCI_BIN)
else
	@echo 1 >/dev/null
endif

# run full lint like in pipeline
.PHONY: lint
lint: install-linter	
	@$(GOLANGCI_BIN) cache clean && \
	$(GOLANGCI_BIN) run --config=$(CURDIR)/.golangci.yaml -v $(CURDIR)/...

# install project dependencies
.PHONY: go-deps
go-deps:
	$(info Install dependencies...)
	@go mod tidy && go mod vendor && go mod verify

.PHONY: test
test:	
	@echo running tests... &&\
	go clean -testcache &&\
	go test -coverprofile=cover.txt -v  ./... &&\
	go tool cover -html=cover.txt -o cover.html &&\
	rm ./cover.txt &&\
	echo "-=OK=-"





