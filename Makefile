BIN := bin/infinita
VERSION ?= dev
LDFLAGS := -ldflags "-X github.com/handiism/infinita/internal/version.Version=$(VERSION)"
M := migrate
DIR := internal/infrastructure/database/sqlite/migrations
DB ?= /tmp/infinita-dev.db

.PHONY: build run test cover lint tidy ci mg-create mgu mgd mgd-all mgf mgv sg

build:
	go build $(LDFLAGS) -o $(BIN) ./cmd/cli

run:
	go run ./cmd/cli $(ARGS)

test:
	go test ./...

cover:
	go test -cover ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

ci: lint test build

mg-create:
	@test -n "$(name)" || (echo "Usage: make mg-create name=<desc>" && exit 1)
	$(M) create -ext sql -dir "$(DIR)" -seq -digits 6 "$(name)"

mgu:
	$(M) -path "$(DIR)" -database "sqlite3://$(DB)" up

mgd:
	$(M) -path "$(DIR)" -database "sqlite3://$(DB)" down 1

mgd-all:
	$(M) -path "$(DIR)" -database "sqlite3://$(DB)" down -all

mgf:
	@test -n "$(version)" || (echo "Usage: make mgf version=<ver>" && exit 1)
	$(M) -path "$(DIR)" -database "sqlite3://$(DB)" force $(version)

mgv:
	$(M) -path "$(DIR)" -database "sqlite3://$(DB)" version

sg:
	sqlc generate
