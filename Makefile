PROJECTNAME=$(shell basename "$(PWD)")

# Go related variables.
# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: setup
## setup: Setup installes dependencies
setup:
	go mod tidy -compat=1.19

.PHONY: test
## test: Runs go test with default values
test:
	go test -v -race -count=1  ./...

.PHONY: build
## build: Builds a beta version of gotoaws
build:
	go build -o dist/

.PHONY: ci
## ci: Run all the tests and code checks
ci: build test

.PHONY: zip-slip
## zip-slip: Runs zipbomb zip-slip
zip-slip:
	go run -race main.go zip-slip --zip-slip "../script.sh" --verify

.PHONY: fast-run
## fast-run: Runs zipbomb
fast-run:
	go run -race main.go overlap -N 2000 --verify

.PHONY: run
## run: Runs zipbomb
run:
	go run -race main.go overlap -N 2000 -R 200000000 --verify

.PHONY: help
## help: Prints this help message
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo