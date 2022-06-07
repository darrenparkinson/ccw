include .envrc


# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run -race ./cmd/api -env development

## run/cli: run the cmd/cli application
.PHONY: run/cli
run/cli:
	@go run -race ./cmd/cli -env development

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	# requires staticcheck.io from go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

## coverage: create coverage report
.PHONY: coverage
coverage:
	@echo 'Creating coverage report...'
	go test -coverprofile=./coverage/profile.out ./...
	go tool cover -func=./coverage/profile.out -o ./coverage/coverage.txt
	go tool cover -html=./coverage/profile.out -o ./coverage/coverage.html

# ==================================================================================== #
# BUILD
# ==================================================================================== #

# Ensure date supports the option
# current_time = $(shell date --iso-8601=seconds)
# current_time = $(shell date)
git_description = $(shell git describe --always --dirty --tags --long)
# -s used to reduce binary size by removing debug info
# linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'
linker_flags = '-s -X main.version=${git_description}'

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/ccw-api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/ccw-api ./cmd/api
	GOOS=windows GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/windows_amd64/ccw-api ./cmd/api
	GOOS=darwin GOARCH=arm64 go build -ldflags=${linker_flags} -o=./bin/darwin_arm64/ccw-api ./cmd/api

## build/cli: build the cmd/cli application
.PHONY: build/cli
build/cli:
	@echo 'Building cmd/cli...'
	go build -ldflags=${linker_flags} -o=./bin/ccw ./cmd/cli
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/ccw ./cmd/cli
	GOOS=windows GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/windows_amd64/ccw ./cmd/cli
	GOOS=darwin GOARCH=arm64 go build -ldflags=${linker_flags} -o=./bin/darwin_arm64/ccw ./cmd/cli
