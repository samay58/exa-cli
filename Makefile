GO ?= go

.PHONY: build install fmt test test-cover test-live docs battle-test battle-test-live

build:
	mkdir -p ./bin
	$(GO) build -o ./bin/exa-cli ./cmd/exa-cli

install: build
	cp ./bin/exa-cli /usr/local/bin/exa-cli

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

test-cover:
	$(GO) test -cover ./...

test-live:
	EXA_LIVE_TESTS=1 $(GO) test ./internal/exa -run Live -count=1

docs:
	$(GO) run ./cmd/exa-cli gen docs > docs/REFERENCE.md

battle-test:
	./scripts/battle-test.sh

battle-test-live:
	EXA_BATTLE_TEST_LIVE=1 ./scripts/battle-test.sh
