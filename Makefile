.DEFAULT_GOAL := help

.PHONY: help
help: ## Outputs the help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Compiles the application
	go build -race -o dropbox-auto-cleaner

.PHONY: vet
vet: ## Runs go vet
	go vet ./...

.PHONY: run
run: build ## Compiles and starts the application
	./dropbox-auto-cleaner

.PHONY: build-docker
build-docker: ## Builds the docker image
	docker build -t dropbox-auto-cleaner:`git log --pretty=format:'%H' -n 1` .