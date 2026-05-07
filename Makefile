#!/usr/bin/make -f

# set variables
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif
LEDGER_ENABLED ?= true
COSMOS_SDK_VERSION := $(shell go list -m github.com/cosmos/cosmos-sdk | sed 's:.* ::')
CMT_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
DOCKER := $(shell which docker)

# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
	ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
   endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=safrochain \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=safrochaind \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(CMT_VERSION)

ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################

# ANSI styling helpers (degrade gracefully when stdout is not a TTY).
C_RESET   := \033[0m
C_BOLD    := \033[1m
C_DIM     := \033[2m
C_CYAN    := \033[38;5;51m
C_BLUE    := \033[38;5;75m
C_GREEN   := \033[38;5;82m
C_YELLOW  := \033[38;5;221m
C_MAGENTA := \033[38;5;213m
C_GREY    := \033[38;5;245m

# Multi-line SAFROCHAIN ASCII banner. Exported so recipes can `printf "$$BANNER"`.
define BANNER

 #####    #####   #######  ######    #####    ######  ##   ##   #####   #####  ##   ##
##       ##   ##  ##       ##   ##  ##   ##  ##       ##   ##  ##   ##   ###   ###  ##
 #####   #######  ######   ######   ##   ##  ##       #######  #######   ###   ## # ##
     ##  ##   ##  ##       ##  ##   ##   ##  ##       ##   ##  ##   ##   ###   ##  ###
 #####   ##   ##  ##       ##   ##   #####    ######  ##   ##  ##   ##  #####  ##   ##

endef
export BANNER

# Pretty-print the build/install summary. Inputs: $(SUMMARY_KIND), $(SUMMARY_BIN), $(SUMMARY_RUN).
print-summary:
	@printf '\n$(C_CYAN)%s$(C_RESET)' "$$BANNER"
	@printf '$(C_DIM)        Sovereign blockchain · powered by the Cosmos SDK$(C_RESET)\n\n'
	@printf '$(C_BOLD)┌──────────────────────────────────────────────────────────────────$(C_RESET)\n'
	@printf '$(C_BOLD)│$(C_RESET)   $(C_MAGENTA)$(C_BOLD)✨  %s COMPLETE$(C_RESET)\n' "$(SUMMARY_KIND)"
	@printf '$(C_BOLD)├──────────────────────────────────────────────────────────────────$(C_RESET)\n'
	@printf '$(C_BOLD)│$(C_RESET)   $(C_GREEN)●$(C_RESET)  safrochaind   $(C_YELLOW)%s$(C_RESET)\n' "$(VERSION)"
	@printf '$(C_BOLD)│$(C_RESET)   $(C_GREEN)●$(C_RESET)  Cosmos SDK    $(C_YELLOW)%s$(C_RESET)\n' "$(COSMOS_SDK_VERSION)"
	@printf '$(C_BOLD)│$(C_RESET)   $(C_GREEN)●$(C_RESET)  CometBFT      $(C_YELLOW)%s$(C_RESET)\n' "$(CMT_VERSION)"
	@printf '$(C_BOLD)│$(C_RESET)   $(C_GREEN)●$(C_RESET)  Go runtime    $(C_YELLOW)%s$(C_RESET)\n' "$$(go version 2>/dev/null | awk '{print $$3, $$4}')"
	@printf '$(C_BOLD)│$(C_RESET)   $(C_GREEN)●$(C_RESET)  Build tags    $(C_GREY)%s$(C_RESET)\n' "$(build_tags_comma_sep)"
	@printf '$(C_BOLD)│$(C_RESET)   $(C_GREEN)●$(C_RESET)  Commit        $(C_GREY)%s$(C_RESET)\n' "$$(echo $(COMMIT) | cut -c1-12)"
	@if [ -n "$(SUMMARY_BIN)" ] && [ -e "$(SUMMARY_BIN)" ]; then \
		size=$$(du -h "$(SUMMARY_BIN)" 2>/dev/null | awk '{print $$1}'); \
		printf '$(C_BOLD)│$(C_RESET)   $(C_GREEN)●$(C_RESET)  Binary        $(C_BLUE)%s$(C_RESET) $(C_DIM)(%s)$(C_RESET)\n' "$(SUMMARY_BIN)" "$$size"; \
	elif [ -n "$(SUMMARY_BIN)" ]; then \
		printf '$(C_BOLD)│$(C_RESET)   $(C_GREEN)●$(C_RESET)  Binary        $(C_BLUE)%s$(C_RESET)\n' "$(SUMMARY_BIN)"; \
	fi
	@printf '$(C_BOLD)└──────────────────────────────────────────────────────────────────$(C_RESET)\n\n'
	@printf '  $(C_DIM)→ Docs:     $(C_RESET) $(C_BLUE)https://docs.safrochain.com$(C_RESET)\n\n'

verify:
	@printf '$(C_CYAN)🔎 Verifying dependencies ...$(C_RESET)\n'
	@go mod verify > /dev/null 2>&1
	@go mod tidy
	@printf '$(C_GREEN)✅ Verified dependencies successfully$(C_RESET)\n\n'

go-cache: verify
	@printf '$(C_CYAN)📥 Downloading and caching dependencies ...$(C_RESET)\n'
	@go mod download
	@printf '$(C_GREEN)✅ Downloaded and cached dependencies successfully$(C_RESET)\n\n'

install: go-cache
	@printf '$(C_CYAN)🔄 Installing safrochaind ...$(C_RESET)\n'
	@go install $(BUILD_FLAGS) -mod=readonly ./cmd/safrochaind
	@mkdir -p ./go/bin
	@INSTALL_DIR="$$(go env GOBIN)"; \
	[ -z "$$INSTALL_DIR" ] && INSTALL_DIR="$$(go env GOPATH)/bin"; \
	cp "$$INSTALL_DIR/safrochaind" ./go/bin/safrochaind || true; \
	printf '$(C_GREEN)✅ Installed safrochaind successfully$(C_RESET)\n'; \
	$(MAKE) --no-print-directory print-summary \
		SUMMARY_KIND="INSTALL" \
		SUMMARY_BIN="$$INSTALL_DIR/safrochaind" \
		SUMMARY_RUN="safrochaind"

build: go-cache
	@printf '$(C_CYAN)🔄 Building safrochaind ...$(C_RESET)\n'
	@if [ "$(OS)" = "Windows_NT" ]; then \
		GOOS=windows GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o bin/safrochaind.exe ./cmd/safrochaind; \
	else \
		go build -mod=readonly $(BUILD_FLAGS) -o bin/safrochaind ./cmd/safrochaind; \
	fi
	@printf '$(C_GREEN)✅ Built safrochaind successfully$(C_RESET)\n'
	@if [ "$(OS)" = "Windows_NT" ]; then \
		$(MAKE) --no-print-directory print-summary SUMMARY_KIND="BUILD" SUMMARY_BIN="./bin/safrochaind.exe" SUMMARY_RUN="./bin/safrochaind.exe"; \
	else \
		$(MAKE) --no-print-directory print-summary SUMMARY_KIND="BUILD" SUMMARY_BIN="./bin/safrochaind" SUMMARY_RUN="./bin/safrochaind"; \
	fi

test-node:
	CHAIN_ID="local-1" HOME_DIR="~/.safrochain" TIMEOUT_COMMIT="500ms" CLEAN=true sh scripts/test_node.sh

.PHONY: verify go-cache install build test-node print-summary

###############################################################################
###                                 Tooling                                 ###
###############################################################################

gofumpt=mvdan.cc/gofumpt
gofumpt_version=v0.8.0

golangci_lint=github.com/golangci/golangci-lint/v2/cmd/golangci-lint
golangci_lint_version=v2.1.6

install-format:
	@echo "🔄 - Installing gofumpt $(gofumpt_version)..."
	@go install $(gofumpt)@$(gofumpt_version)
	@echo "✅ - Installed gofumpt successfully!"
	@echo ""

install-lint:
	@echo "🔄 - Installing golangci-lint $(golangci_lint_version)..."
	@go install $(golangci_lint)@$(golangci_lint_version)
	@echo "✅ - Installed golangci-lint successfully!"
	@echo ""

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		INSTALLED=$$(golangci-lint version | head -n1 | awk '{print $$4}'); \
		echo "Detected golangci-lint $$INSTALLED, required $(golangci_lint_version)"; \
		if [ "$$(printf '%s\n' "$(golangci_lint_version)" "$$INSTALLED" | sort -V | head -n1)" != "$(golangci_lint_version)" ]; then \
	   	echo "Updating golangci-lint..."; \
	   	$(MAKE) install-lint; \
		fi; \
	else \
		echo "golangci-lint not found; installing..."; \
		$(MAKE) install-lint; \
	fi
	@echo "🔄 - Linting code..."
	@golangci-lint run
	@echo "✅ - Linted code successfully!"

format:
	@if command -v gofumpt >/dev/null 2>&1; then \
		INSTALLED=$$(go version -m $$(command -v gofumpt) | awk '$$1=="mod" {print $$3; exit}'); \
		echo "Detected gofumpt $$INSTALLED, required $(gofumpt_version)"; \
		if [ "$$(printf '%s\n' "$(gofumpt_version)" "$$INSTALLED" | sort -V | head -n1)" != "$(gofumpt_version)" ]; then \
	   	echo "Updating gofumpt..."; \
	   	$(MAKE) install-format; \
		fi; \
	else \
		echo "gofumpt not found; installing..."; \
		$(MAKE) install-format; \
	fi
	@echo "🔄 - Formatting code..."
	@gofumpt -l -w .
	@echo "✅ - Formatted code successfully!"

.PHONY: install-format format install-lint lint

###############################################################################
###                             e2e interchain test                         ###
###############################################################################

ictest-basic: rm-testcache
	cd interchaintest && go test -race -v -run TestBasicsafrochainStart .

ictest-statesync: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainStateSync .

ictest-ibchooks: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainIBCHooks .

ictest-tokenfactory: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainTokenFactory .

ictest-feeshare: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainFeeShare .

ictest-pfm: rm-testcache
	cd interchaintest && go test -race -v -run TestPacketForwardMiddlewareRouter .

ictest-globalfee: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainGlobalFee .

ictest-upgrade: rm-testcache
	cd interchaintest && go test -race -v -run TestBasicsafrochainUpgrade .

ictest-ibc: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainGaiaIBCTransfer .

ictest-unity-deploy: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainUnityContractDeploy .

ictest-drip: rm-testcache
	cd interchaintest && go test -race -v -run Testsafrochaindrip .

ictest-feepay: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainFeePay .

ictest-burn: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainBurnModule .

ictest-cwhooks: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainCwHooks .

ictest-clock: rm-testcache
	cd interchaintest && go test -race -v -run TestsafrochainClock .

ictest-gov-fix: rm-testcache
	cd interchaintest && go test -race -v -run TestFixRemovedMsgTypeQueryPanic .

rm-testcache:
	go clean -testcache

.PHONY: ictest-basic ictest-statesync ictest-ibchooks ictest-tokenfactory ictest-feeshare ictest-pfm ictest-globalfee ictest-upgrade ictest-upgrade-local ictest-ibc ictest-unity-deploy ictest-unity-gov ictest-drip ictest-burn ictest-feepay ictest-cwhooks ictest-clock ictest-gov-fix rm-testcache

###############################################################################
###                                  heighliner                             ###
###############################################################################

heighliner=github.com/strangelove-ventures/heighliner
heighliner_version=v1.7.2

install-heighliner:
	@if ! command -v heighliner > /dev/null; then \
   	echo "🔄 - Installing heighliner $(heighliner_version)..."; \
      go install $(heighliner)@$(heighliner_version); \
		echo "✅ - Installed heighliner successfully!"; \
		echo ""; \
   fi

local-image: install-heighliner
	@echo "🔄 - Building Docker Image..."
	heighliner build --chain safrochain --local -f ./chains.yaml
	@echo "✅ - Built Docker Image successfully!"

.PHONY: install-heighliner local-image

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.17.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace -v /var/run/docker.sock:/var/run/docker.sock --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen proto-gen-2 proto-swagger-gen

proto-gen:
	@echo "🛠️ - Generating Protobuf"
	@$(protoImage) sh ./scripts/protoc/protocgen.sh
	@echo "✅ - Generated Protobuf successfully!"

proto-gen-2:
	@echo "🛠️ - Generating Protobuf v2"
	@$(protoImage) sh ./scripts/protoc/protocgen2.sh
	@echo "✅ - Generated Protobuf v2 successfully!"

proto-swagger-gen:
	@echo "📖 - Generating Protobuf Swagger"
	@$(protoImage) sh ./scripts/protoc/protoc-swagger-gen.sh
	@echo "✅ - Generated Protobuf Swagger successfully!"

proto-format:
	@echo "🖊️ - Formatting Protobuf Swagger"
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;
	@echo "✅ - Formatted Protobuf successfully!"

proto-lint:
	@echo "🔎 - Linting Protobuf"
	@$(protoImage) buf lint --error-format=json
	@echo "✅ - Linted Protobuf successfully!"

proto-check-breaking:
	@echo "🔎 - Checking breaking Protobuf changes"
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main
	@echo "✅ - Checked Protobuf changes successfully!"

.PHONY: proto-all proto-gen proto-gen-2 proto-swagger-gen proto-format proto-lint proto-check-breaking
